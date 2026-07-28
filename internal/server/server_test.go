package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/llm"
	"suprie/application_tracker/internal/migration"
	"suprie/application_tracker/internal/repository"
	sqliterepo "suprie/application_tracker/internal/repository/sqlite"
	"suprie/application_tracker/internal/service"
)

func sptr(s string) *string { return &s }

// newTestServer builds a real server against a migrated temp DB and returns it
// plus repos pointing at the same DB, so tests can seed data in-process.
func newTestServer(t *testing.T) (*httptest.Server, repository.JobDescriptionRepository, repository.CompanyRepository) {
	return newTestServerWithLLM(t, nil)
}

// newTestServerWithLLM builds a server whose Deps use the given LLM factory
// (nil = default, which is fine for tests that never trigger an LLM call).
func newTestServerWithLLM(t *testing.T, llmFactory func(string) llm.LLMClient) (*httptest.Server, repository.JobDescriptionRepository, repository.CompanyRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dbPath := t.TempDir() + "/test.db"
	db, err := sqliterepo.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := migration.Run(db, "up"); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	jdRepo := sqliterepo.NewJobDescriptionRepository(db)
	companyRepo := sqliterepo.NewCompanyRepository(db)
	deps := service.NewDeps(jdRepo, companyRepo)
	if llmFactory != nil {
		deps.LLM = llmFactory
	}
	ts := httptest.NewServer(New(deps))
	t.Cleanup(ts.Close)
	return ts, jdRepo, companyRepo
}

func readJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	defer resp.Body.Close()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("unmarshal body %q: %v", string(body), err)
	}
	return m
}

func TestAPI_ListJds_Empty(t *testing.T) {
	ts, _, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/api/jds")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: want 200, got %d", resp.StatusCode)
	}
	m := readJSON(t, resp)
	items, ok := m["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %T", m["items"])
	}
	if len(items) != 0 {
		t.Errorf("want 0 items, got %d", len(items))
	}
}

func TestAPI_JDEndpoints(t *testing.T) {
	ts, jdRepo, _ := newTestServer(t)

	// Seed a JD into the same DB the server reads.
	jd := &domain.JobDescription{
		Company:              sptr("Globex"),
		RoleTitle:            sptr("Backend Engineer"),
		RequirementsJSON:     `{"must_have":["Go"]}`,
		ResponsibilitiesJSON: `["build apis"]`,
		KeywordsJSON:         `["go"]`,
		ParsingWarningJSON:   `[]`,
		CreatedAt:            time.Now(),
	}
	if err := jdRepo.Create(context.Background(), jd); err != nil {
		t.Fatalf("seed jd: %v", err)
	}
	id := jd.ID

	// List returns the seeded record.
	resp, err := http.Get(ts.URL + "/api/jds")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	m := readJSON(t, resp)
	items := m["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}

	// Detail inlines JSON columns as objects/arrays, not escaped strings.
	resp2, err := http.Get(fmt.Sprintf("%s/api/jds/%d", ts.URL, id))
	if err != nil {
		t.Fatalf("detail: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("detail status: got %d", resp2.StatusCode)
	}
	m2 := readJSON(t, resp2)
	if _, ok := m2["requirements"].(map[string]any); !ok {
		t.Errorf("requirements should be a JSON object, got %T (%v)", m2["requirements"], m2["requirements"])
	}

	// Patch status.
	req, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/jds/%d", ts.URL, id), strings.NewReader(`{"status":"applied"}`))
	req.Header.Set("Content-Type", "application/json")
	resp3, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("patch: %v", err)
	}
	m3 := readJSON(t, resp3)
	if m3["status"] != domain.StatusApplied {
		t.Errorf("status after patch: got %v", m3["status"])
	}

	// Apply (idempotent, still 200).
	resp4, err := http.Post(fmt.Sprintf("%s/api/jds/%d/apply", ts.URL, id), "application/json", nil)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if resp4.StatusCode != http.StatusOK {
		t.Errorf("apply status: want 200, got %d", resp4.StatusCode)
	}
	resp4.Body.Close()

	// Unknown id → 404.
	resp5, err := http.Get(ts.URL + "/api/jds/999")
	if err != nil {
		t.Fatalf("nf: %v", err)
	}
	if resp5.StatusCode != http.StatusNotFound {
		t.Errorf("nf status: want 404, got %d", resp5.StatusCode)
	}
	resp5.Body.Close()

	// Bad status → 400.
	reqBad, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/jds/%d", ts.URL, id), strings.NewReader(`{"status":"bogus"}`))
	reqBad.Header.Set("Content-Type", "application/json")
	resp6, err := http.DefaultClient.Do(reqBad)
	if err != nil {
		t.Fatalf("bad patch: %v", err)
	}
	if resp6.StatusCode != http.StatusBadRequest {
		t.Errorf("bad status patch: want 400, got %d", resp6.StatusCode)
	}
	resp6.Body.Close()
}

func TestAPI_CompanyCRUD(t *testing.T) {
	ts, _, _ := newTestServer(t)

	// Create.
	resp, err := http.Post(ts.URL+"/api/companies", "application/json",
		strings.NewReader(`{"name":"Acme Corp","industry":"Manufacturing","website_url":"https://acme.example"}`))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status: want 201, got %d", resp.StatusCode)
	}
	m := readJSON(t, resp)
	if m["name"] != "Acme Corp" {
		t.Errorf("name: got %v", m["name"])
	}
	if m["normalized_name"] != "acme" {
		t.Errorf("normalized_name: got %v", m["normalized_name"])
	}
	id := int(m["id"].(float64))

	// Duplicate (different display name, same normalized) → 409.
	resp2, err := http.Post(ts.URL+"/api/companies", "application/json", strings.NewReader(`{"name":"Acme Inc."}`))
	if err != nil {
		t.Fatalf("dup: %v", err)
	}
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("dup status: want 409, got %d", resp2.StatusCode)
	}
	resp2.Body.Close()

	// Search.
	resp3, err := http.Get(ts.URL + "/api/companies?q=acme")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	m3 := readJSON(t, resp3)
	items := m3["items"].([]any)
	if len(items) != 1 {
		t.Errorf("search items: want 1, got %d", len(items))
	}

	// Update.
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/companies/%d", ts.URL, id), strings.NewReader(`{"industry":"Tech"}`))
	req.Header.Set("Content-Type", "application/json")
	resp4, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	m4 := readJSON(t, resp4)
	if m4["industry"] != "Tech" {
		t.Errorf("industry: got %v", m4["industry"])
	}
}

func TestAPI_RootServesIndex(t *testing.T) {
	ts, _, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("root status: got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	// Robust to both the placeholder and a real built index.html.
	if !strings.Contains(string(body), "<html") {
		t.Errorf("expected an HTML index document, got %q", string(body))
	}
}

func TestAPI_SPAFallback(t *testing.T) {
	ts, _, _ := newTestServer(t)

	// An unknown non-API route must fall back to index.html (200), not 404.
	resp, err := http.Get(ts.URL + "/jds/5")
	if err != nil {
		t.Fatalf("spa route: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("spa fallback status: want 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// stubLLM is a fixed-response LLMClient for testing the async pipeline without
// hitting a real model.
type stubLLM struct{ resp string }

func (s stubLLM) Generate(ctx context.Context, prompt string, _ *llm.ResponseFormat) (string, error) {
	return s.resp, nil
}

func stubFactory(resp string) func(string) llm.LLMClient {
	c := stubLLM{resp: resp}
	return func(string) llm.LLMClient { return c }
}

func TestAPI_AsyncParseJD(t *testing.T) {
	resp := `{"company":"Stub Co","role_title":"Engineer","requirements":{"must_have":["Go"],"nice_have":[]},"responsibilities":["build"],"keywords":["go"],"parsing_warnings":[]}`
	ts, _, _ := newTestServerWithLLM(t, stubFactory(resp))

	// POST parse → 202 + task_id.
	r, err := http.Post(ts.URL+"/api/jds", "application/json", strings.NewReader(`{"text":"some jd text"}`))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if r.StatusCode != http.StatusAccepted {
		t.Fatalf("parse status: want 202, got %d", r.StatusCode)
	}
	m := readJSON(t, r)
	taskID, ok := m["task_id"].(string)
	if !ok {
		t.Fatalf("no task_id in response: %v", m)
	}

	// Poll /api/tasks/:id until done.
	var jdID int
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rt, err := http.Get(ts.URL + "/api/tasks/" + taskID)
		if err != nil {
			t.Fatalf("poll: %v", err)
		}
		mt := readJSON(t, rt)
		switch mt["state"] {
		case "done":
			res := mt["result"].(map[string]any)
			jdID = int(res["jd_id"].(float64))
		case "failed":
			t.Fatalf("parse task failed: %v", mt["error"])
		default:
			time.Sleep(10 * time.Millisecond)
			continue
		}
		break
	}
	if jdID == 0 {
		t.Fatal("parse task did not complete in time")
	}

	// The parsed JD is persisted and retrievable.
	rd, err := http.Get(fmt.Sprintf("%s/api/jds/%d", ts.URL, jdID))
	if err != nil {
		t.Fatalf("get jd: %v", err)
	}
	md := readJSON(t, rd)
	if md["company"] != "Stub Co" {
		t.Errorf("company: want Stub Co, got %v", md["company"])
	}

	// Unknown task → 404.
	rnf, _ := http.Get(ts.URL + "/api/tasks/nope")
	if rnf.StatusCode != http.StatusNotFound {
		t.Errorf("nf task status: want 404, got %d", rnf.StatusCode)
	}
	rnf.Body.Close()
}
