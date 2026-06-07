package jdextractor

import (
	"testing"
)

func TestBuildJSONSchema_TypeIsJSONSchema(t *testing.T) {
	schema := BuildJSONSchema()

	if schema.Type != "json_schema" {
		t.Errorf("Type = %q, want %q", schema.Type, "json_schema")
	}
}

func TestBuildJSONSchema_NameIsSet(t *testing.T) {
	schema := BuildJSONSchema()

	if schema.JSONSchema.Name != "parsed_job_description" {
		t.Errorf("Name = %q, want %q", schema.JSONSchema.Name, "parsed_job_description")
	}
}

func TestBuildJSONSchema_RootTypeIsObject(t *testing.T) {
	schema := BuildJSONSchema()

	rootType, ok := schema.JSONSchema.Schema["type"].(string)
	if !ok {
		t.Fatalf("root schema missing 'type' field")
	}
	if rootType != "object" {
		t.Errorf("root type = %q, want %q", rootType, "object")
	}
}

func TestBuildJSONSchema_AllRequiredFieldsHaveProperties(t *testing.T) {
	schema := BuildJSONSchema()

	required, ok := schema.JSONSchema.Schema["required"].([]string)
	if !ok {
		t.Fatalf("root schema missing 'required' array")
	}

	properties, ok := schema.JSONSchema.Schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("root schema missing 'properties' map")
	}

	for _, field := range required {
		if _, exists := properties[field]; !exists {
			t.Errorf("field %q is required but missing from properties", field)
		}
	}
}

func TestBuildJSONSchema_NoExtraRequiredFields(t *testing.T) {
	schema := BuildJSONSchema()

	properties, _ := schema.JSONSchema.Schema["properties"].(map[string]any)
	required, _ := schema.JSONSchema.Schema["required"].([]string)

	requiredSet := make(map[string]bool, len(required))
	for _, r := range required {
		requiredSet[r] = true
	}

	for prop := range properties {
		if !requiredSet[prop] {
			// Optional properties are fine, just note them
			t.Logf("property %q is defined but not required (this may be intentional)", prop)
		}
	}
}

func TestBuildJSONSchema_RequirementsHasRequiredSubFields(t *testing.T) {
	schema := BuildJSONSchema()

	properties, _ := schema.JSONSchema.Schema["properties"].(map[string]any)
	reqs, ok := properties["requirements"].(map[string]any)
	if !ok {
		t.Fatalf("requirements property missing or wrong type")
	}

	subRequired, ok := reqs["required"].([]string)
	if !ok {
		t.Fatalf("requirements.required missing or wrong type")
	}

	want := []string{"must_have", "nice_have"}
	if len(subRequired) != len(want) {
		t.Errorf("requirements.required has %d fields, want %d", len(subRequired), len(want))
	}

	for _, field := range want {
		found := false
		for _, r := range subRequired {
			if r == field {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("requirements.required missing %q", field)
		}
	}
}

func TestBuildJSONSchema_ArrayFieldsHaveStringItems(t *testing.T) {
	schema := BuildJSONSchema()

	properties, _ := schema.JSONSchema.Schema["properties"].(map[string]any)

	arrayFields := []string{"responsibilities", "parsing_warnings", "keywords"}
	for _, field := range arrayFields {
		prop, ok := properties[field].(map[string]any)
		if !ok {
			t.Errorf("field %q missing or wrong type", field)
			continue
		}

		typ, _ := prop["type"].(string)
		if typ != "array" {
			t.Errorf("%s.type = %q, want %q", field, typ, "array")
		}

		items, ok := prop["items"].(map[string]any)
		if !ok {
			t.Errorf("%s.items missing", field)
			continue
		}
		if items["type"] != "string" {
			t.Errorf("%s.items.type = %v, want %q", field, items["type"], "string")
		}
	}
}

func TestBuildJSONSchema_StringOrNullFieldsUseTuple(t *testing.T) {
	schema := BuildJSONSchema()

	properties, _ := schema.JSONSchema.Schema["properties"].(map[string]any)

	nullableFields := []string{"company", "role_title", "seniority", "location", "work_arrangement", "employment_type"}
	for _, field := range nullableFields {
		prop, ok := properties[field].(map[string]any)
		if !ok {
			t.Errorf("field %q missing or wrong type", field)
			continue
		}

		types, ok := prop["type"].([]string)
		if !ok {
			t.Errorf("%s.type should be a []string (tuple), got %T", field, prop["type"])
			continue
		}

		hasString := false
		hasNull := false
		for _, t := range types {
			if t == "string" {
				hasString = true
			}
			if t == "null" {
				hasNull = true
			}
		}
		if !hasString || !hasNull {
			t.Errorf("%s.type = %v, want [string, null]", field, types)
		}
	}
}
