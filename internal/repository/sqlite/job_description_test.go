package sqlite

import (
	"context"
	"fmt"
	"testing"
	"time"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/migration"
)

func openTestDB(t *testing.T) *JobDescriptionRepository {
	t.Helper()

	dbPath := t.TempDir() + "/test.db"

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := migration.Run(db, "up"); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	return NewJobDescriptionRepository(db)
}

func ptr(s string) *string { return &s }

func TestJobDescriptionRepository_CreateAndGetByID(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	jd := &domain.JobDescription{
		Company:              ptr("Acme Corp"),
		RoleTitle:            ptr("Backend Engineer"),
		Seniority:            ptr("Senior"),
		EmploymentType:       ptr("Full-time"),
		WorkArrangement:      ptr("Remote"),
		Location:             ptr("Jakarta"),
		RequirementsJSON:     `{"must_have":["Go","SQL"]}`,
		ResponsibilitiesJSON: `["Build APIs","Review code"]`,
		KeywordsJSON:         `["golang","postgres"]`,
		ParsingWarningJSON:   `[]`,
		CreatedAt:            time.Now(),
	}

	if err := repo.Create(ctx, jd); err != nil {
		t.Fatalf("create: %v", err)
	}

	if jd.ID == 0 {
		t.Fatal("expected Id to be set after create, got 0")
	}

	got, err := repo.GetByID(ctx, jd.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}

	if *got.Company != *jd.Company {
		t.Errorf("Company: want %q, got %q", *jd.Company, *got.Company)
	}
	if *got.Seniority != "Senior" {
		t.Errorf("Seniority: want Senior, got %q", *got.Seniority)
	}
	if got.RequirementsJSON != jd.RequirementsJSON {
		t.Errorf("RequirementsJson mismatch")
	}
}

func TestJobDescriptionRepository_GetByID_NotFound(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestJobDescriptionRepository_GetAll(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	for _, company := range []string{"Alpha", "Beta", "Gamma"} {
		if err := repo.Create(ctx, &domain.JobDescription{
			Company:              ptr(company),
			RoleTitle:            ptr("Engineer"),
			RequirementsJSON:     "{}",
			ResponsibilitiesJSON: "[]",
			KeywordsJSON:         "[]",
			ParsingWarningJSON:   "[]",
			CreatedAt:            time.Now(),
		}); err != nil {
			t.Fatalf("create %s: %v", company, err)
		}
	}

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}

	if len(all) != 3 {
		t.Errorf("want 3 records, got %d", len(all))
	}

	// Verify descending order by created_at.
	if all[0].ID < all[2].ID {
		t.Error("expected descending order by created_at")
	}
}

func TestJobDescriptionRepository_GetAll_Empty(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}

	if len(all) != 0 {
		t.Errorf("want 0 records, got %d", len(all))
	}
}

func TestJobDescriptionRepository_Create_DefaultsToDraft(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	jd := &domain.JobDescription{
		Company:              ptr("Test"),
		RequirementsJSON:     "{}",
		ResponsibilitiesJSON: "[]",
		KeywordsJSON:         "[]",
		ParsingWarningJSON:   "[]",
		CreatedAt:            time.Now(),
	}

	if err := repo.Create(ctx, jd); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, _ := repo.GetByID(ctx, jd.ID)
	if got.Status != domain.StatusDraft {
		t.Errorf("want status Draft, got %q", got.Status)
	}
}

func TestJobDescriptionRepository_List_WithFilter(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	// Create two drafts and one rejected.
	for i, status := range []string{domain.StatusDraft, domain.StatusDraft, domain.StatusRejected} {
		jd := &domain.JobDescription{
			Company:              ptr(fmt.Sprintf("Company-%d", i)),
			Status:               status,
			RequirementsJSON:     "{}",
			ResponsibilitiesJSON: "[]",
			KeywordsJSON:         "[]",
			ParsingWarningJSON:   "[]",
			CreatedAt:            time.Now(),
		}
		if err := repo.Create(ctx, jd); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	all, err := repo.List(ctx, "")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("want 3 total, got %d", len(all))
	}

	drafts, err := repo.List(ctx, domain.StatusDraft)
	if err != nil {
		t.Fatalf("list drafts: %v", err)
	}
	if len(drafts) != 2 {
		t.Errorf("want 2 drafts, got %d", len(drafts))
	}

	rejected, err := repo.List(ctx, domain.StatusRejected)
	if err != nil {
		t.Fatalf("list rejected: %v", err)
	}
	if len(rejected) != 1 {
		t.Errorf("want 1 rejected, got %d", len(rejected))
	}
}

func TestJobDescriptionRepository_UpdateStatus(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	jd := &domain.JobDescription{
		Company:              ptr("Update Me"),
		RequirementsJSON:     "{}",
		ResponsibilitiesJSON: "[]",
		KeywordsJSON:         "[]",
		ParsingWarningJSON:   "[]",
		CreatedAt:            time.Now(),
	}
	if err := repo.Create(ctx, jd); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, jd.ID, domain.StatusApplied); err != nil {
		t.Fatalf("update status: %v", err)
	}

	got, _ := repo.GetByID(ctx, jd.ID)
	if got.Status != domain.StatusApplied {
		t.Errorf("want status Applied, got %q", got.Status)
	}
	if got.AppliedAt == nil {
		t.Error("expected AppliedAt to be set when status changes to Applied")
	}
}

func TestJobDescriptionRepository_UpdateFitScore(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	jd := &domain.JobDescription{
		Company:              ptr("Fit Match Co"),
		RequirementsJSON:     "{}",
		ResponsibilitiesJSON: "[]",
		KeywordsJSON:         "[]",
		ParsingWarningJSON:   "[]",
		CreatedAt:            time.Now(),
	}
	if err := repo.Create(ctx, jd); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := repo.UpdateFitScore(ctx, jd.ID, 85, "Good fit overall"); err != nil {
		t.Fatalf("update fit score: %v", err)
	}

	got, _ := repo.GetByID(ctx, jd.ID)
	if got.Status != domain.StatusFitMatch {
		t.Errorf("want status Fit match, got %q", got.Status)
	}
	if got.FitScore == nil || *got.FitScore != 85 {
		t.Errorf("want fit_score 85, got %v", got.FitScore)
	}
	if got.FitSummary == nil || *got.FitSummary != "Good fit overall" {
		t.Errorf("want fit_summary 'Good fit overall', got %v", got.FitSummary)
	}
}

func TestJobDescriptionRepository_NullableFields(t *testing.T) {
	repo := openTestDB(t)
	ctx := context.Background()

	// Create with nil optional fields.
	jd := &domain.JobDescription{
		RequirementsJSON:     "{}",
		ResponsibilitiesJSON: "[]",
		KeywordsJSON:         "[]",
		ParsingWarningJSON:   "[]",
		CreatedAt:            time.Now(),
	}

	if err := repo.Create(ctx, jd); err != nil {
		t.Fatalf("create with nil fields: %v", err)
	}

	got, err := repo.GetByID(ctx, jd.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}

	if got.Company != nil {
		t.Errorf("Company should be nil, got %q", *got.Company)
	}
	if got.Location != nil {
		t.Errorf("Location should be nil, got %q", *got.Location)
	}
}
