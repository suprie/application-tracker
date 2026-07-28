package server

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"suprie/application_tracker/internal/service"
)

// newHTTPTestServer builds a server directly from the given deps, bypassing
// the DB-backed newTestServer helper — profile upload needs neither repo.
func newHTTPTestServer(t *testing.T, deps service.Deps) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(New(deps))
	t.Cleanup(ts.Close)
	return ts
}

// writeFakeReader creates a shell script standing in for the Rust PDF
// extractor binary, echoing the given JSON on stdout.
func writeFakeReader(t *testing.T, jsonOutput string) string {
	t.Helper()
	dir := t.TempDir()
	var path, content string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "fake-reader.bat")
		content = fmt.Sprintf("@echo off\r\necho %s\r\n", jsonOutput)
	} else {
		path = filepath.Join(dir, "fake-reader.sh")
		content = fmt.Sprintf("#!/bin/sh\necho '%s'\n", jsonOutput)
	}
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		t.Fatalf("write fake reader: %v", err)
	}
	return path
}

func TestAPI_UploadCV_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := service.NewDeps(nil, nil)
	ts := newHTTPTestServer(t, deps)

	resp, err := http.Post(ts.URL+"/api/profile/cv", "application/octet-stream", nil)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: want 400, got %d", resp.StatusCode)
	}
}

func TestAPI_UploadCV_RejectsNonPDF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := service.NewDeps(nil, nil)
	ts := newHTTPTestServer(t, deps)

	body, contentType := multipartCV(t, "resume.txt", "not a pdf")
	resp, err := http.Post(ts.URL+"/api/profile/cv", contentType, body)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: want 400, got %d", resp.StatusCode)
	}
}

func TestAPI_UploadCV_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fakeReader := writeFakeReader(t, `{"text": "Jane Doe, Software Engineer"}`)
	profilePath := filepath.Join(t.TempDir(), "master_profile.yaml")

	deps := service.NewDeps(nil, nil)
	deps.ReaderBinPath = fakeReader
	deps.ProfilePath = profilePath
	deps.LLM = stubFactory("name: Jane Doe\ntitle: Software Engineer\n")

	ts := newHTTPTestServer(t, deps)

	body, contentType := multipartCV(t, "resume.pdf", "%PDF-1.4 fake bytes")
	resp, err := http.Post(ts.URL+"/api/profile/cv", contentType, body)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: want 202, got %d", resp.StatusCode)
	}
	m := readJSON(t, resp)
	taskID, ok := m["task_id"].(string)
	if !ok {
		t.Fatalf("no task_id in response: %v", m)
	}

	done := false
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		rt, err := http.Get(ts.URL + "/api/tasks/" + taskID)
		if err != nil {
			t.Fatalf("poll: %v", err)
		}
		mt := readJSON(t, rt)
		switch mt["state"] {
		case "done":
			done = true
		case "failed":
			t.Fatalf("upload task failed: %v", mt["error"])
		default:
			time.Sleep(10 * time.Millisecond)
			continue
		}
		break
	}
	if !done {
		t.Fatal("upload task did not complete in time")
	}

	content, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("read profile: %v", err)
	}
	if !bytes.Contains(content, []byte("Jane Doe")) {
		t.Errorf("profile content = %q, want it to contain %q", content, "Jane Doe")
	}
}

// multipartCV builds a multipart/form-data body with a single "cv" file field.
func multipartCV(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("cv", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	return &buf, w.FormDataContentType()
}
