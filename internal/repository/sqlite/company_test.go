package sqlite

import (
	"context"
	"strings"
	"testing"

	"suprie/application_tracker/internal/domain"
	"suprie/application_tracker/internal/migration"
)

func openCompanyTestDB(t *testing.T) *CompanyRepository {
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

	return NewCompanyRepository(db)
}

func TestCompanyRepository_CreateAndGetByID(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	c := &domain.Company{
		Name:            "Acme Corp",
		NormalizedName:  "acme",
		WebsiteURL:      ptr("https://acme.example"),
		Industry:        ptr("Manufacturing"),
		Size:            ptr("500-1000"),
		Country:         ptr("Indonesia"),
		Notes:           ptr("Fast-growing"),
		Source:          ptr("manual"),
		ResearchSummary: ptr("Leader in widgets."),
	}

	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID == 0 {
		t.Fatal("expected ID to be set after create")
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.Name != c.Name {
		t.Errorf("Name: want %q, got %q", c.Name, got.Name)
	}
	if got.NormalizedName != c.NormalizedName {
		t.Errorf("NormalizedName: want %q, got %q", c.NormalizedName, got.NormalizedName)
	}
	if got.WebsiteURL == nil || *got.WebsiteURL != "https://acme.example" {
		t.Errorf("WebsiteURL mismatch: %v", got.WebsiteURL)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Error("expected timestamps to be set")
	}
}

func TestCompanyRepository_GetByID_NotFound(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, 999); err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
}

func TestCompanyRepository_GetByNormalizedName(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	c := &domain.Company{Name: "Google LLC", NormalizedName: "google"}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := repo.GetByNormalizedName(ctx, "google")
	if err != nil {
		t.Fatalf("get by normalized name: %v", err)
	}
	if got.ID != c.ID {
		t.Errorf("ID: want %d, got %d", c.ID, got.ID)
	}

	if _, err := repo.GetByNormalizedName(ctx, "missing"); err == nil {
		t.Fatal("expected error for missing normalized name, got nil")
	}
}

func TestCompanyRepository_Create_DuplicateNormalizedName(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	if err := repo.Create(ctx, &domain.Company{Name: "Apple Inc.", NormalizedName: "apple"}); err != nil {
		t.Fatalf("create first: %v", err)
	}

	if err := repo.Create(ctx, &domain.Company{Name: "Apple", NormalizedName: "apple"}); err == nil {
		t.Fatal("expected error creating duplicate normalized_name, got nil")
	}
}

func TestCompanyRepository_List(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	for _, name := range []string{"Gamma", "Alpha", "Beta"} {
		if err := repo.Create(ctx, &domain.Company{Name: name, NormalizedName: strings.ToLower(name)}); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("want 3, got %d", len(all))
	}
	// Ordered by name ASC.
	if all[0].Name != "Alpha" || all[2].Name != "Gamma" {
		t.Errorf("expected alphabetical order, got %s, %s, %s", all[0].Name, all[1].Name, all[2].Name)
	}
}

func TestCompanyRepository_List_Empty(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	all, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if all == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(all) != 0 {
		t.Errorf("want 0, got %d", len(all))
	}
}

func TestCompanyRepository_Update(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	c := &domain.Company{Name: "Acme", NormalizedName: "acme", Industry: ptr("Old")}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create: %v", err)
	}

	c.Industry = ptr("New")
	c.Notes = ptr("Updated notes")
	if err := repo.Update(ctx, c); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if got.Industry == nil || *got.Industry != "New" {
		t.Errorf("Industry: want New, got %v", got.Industry)
	}
	if got.Notes == nil || *got.Notes != "Updated notes" {
		t.Errorf("Notes mismatch: %v", got.Notes)
	}
}

func TestCompanyRepository_NullableFields(t *testing.T) {
	repo := openCompanyTestDB(t)
	ctx := context.Background()

	c := &domain.Company{Name: "Minimal", NormalizedName: "minimal"}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("create with nil fields: %v", err)
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.WebsiteURL != nil {
		t.Errorf("WebsiteURL should be nil, got %q", *got.WebsiteURL)
	}
	if got.Industry != nil {
		t.Errorf("Industry should be nil, got %q", *got.Industry)
	}
}
