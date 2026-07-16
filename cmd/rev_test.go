package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/experimental-software/gherkin/core"
)

func TestParseSpecSources_File(t *testing.T) {
	docs, err := parseSpecSources("../core/testdata/sample.spec.ts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
}

func TestParseSpecSources_DirectoryRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	sampleContent, err := os.ReadFile("../core/testdata/sample.spec.ts")
	if err != nil {
		t.Fatalf("failed to read sample fixture: %v", err)
	}
	nestedContent, err := os.ReadFile("../core/testdata/nested-describe-blocks.spec.ts")
	if err != nil {
		t.Fatalf("failed to read nested fixture: %v", err)
	}

	sampleTarget := filepath.Join(tmpDir, "sample.spec.ts")
	if err := os.WriteFile(sampleTarget, sampleContent, 0o644); err != nil {
		t.Fatalf("failed to write sample fixture: %v", err)
	}

	nestedDir := filepath.Join(tmpDir, "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create nested directory: %v", err)
	}
	nestedTarget := filepath.Join(nestedDir, "nested-describe-blocks.spec.ts")
	if err := os.WriteFile(nestedTarget, nestedContent, 0o644); err != nil {
		t.Fatalf("failed to write nested fixture: %v", err)
	}

	ignoredFile := filepath.Join(tmpDir, "ignored.ts")
	if err := os.WriteFile(ignoredFile, []byte("describe('Ignored', () => {})"), 0o644); err != nil {
		t.Fatalf("failed to write ignored file: %v", err)
	}

	docs, err := parseSpecSources(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 3 {
		t.Fatalf("expected 3 documents parsed from recursive spec files, got %d", len(docs))
	}
}

func TestParseSpecSources_MissingPath(t *testing.T) {
	_, err := parseSpecSources("testdata/nonexistent-dir")
	if err == nil {
		t.Fatal("expected error for missing source path, got nil")
	}
}

func TestSanitizeFeatureFileBaseName(t *testing.T) {
	got := sanitizeFeatureFileBaseName("User authentication")
	want := "User_authentication"
	if got != want {
		t.Fatalf("expected sanitized base name %q, got %q", want, got)
	}
}

func TestFeatureFilePathPartsFromURI(t *testing.T) {
	relativeDir, baseName := featureFilePathPartsFromURI("src/app/auth.spec.ts/User_authentication")
	if relativeDir != filepath.Join("src", "app", "auth.spec.ts") {
		t.Fatalf("expected relative dir %q, got %q", filepath.Join("src", "app", "auth.spec.ts"), relativeDir)
	}
	if baseName != "User_authentication" {
		t.Fatalf("expected base name %q, got %q", "User_authentication", baseName)
	}
}

func TestWriteFeatureFiles_DiscardsSpecFileName(t *testing.T) {
	docs, err := parseSpecSources("../core/testdata/sample.spec.ts")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}

	targetDir := t.TempDir()
	if err := writeFeatureFiles(docs, targetDir); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	// Given an URI like core/testdata/sample.spec.ts/user-authentication
	// core.CalculateFeaturePath(docs[0].Uri) should give core/testdata/user-authentication
	uri := core.CalculateFeaturePath(docs[0].Uri)
	relativeDir, baseName := featureFilePathPartsFromURI(uri)
	featurePath := filepath.Join(targetDir, filepath.FromSlash(relativeDir), baseName+".feature")

	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		t.Fatalf("expected feature file to be written at %s, but it was not found", featurePath)
	}

	// Verify that the path does NOT contain .spec.ts
	if strings.Contains(featurePath, ".spec.ts") {
		t.Fatalf("expected feature path %q to NOT contain '.spec.ts'", featurePath)
	}
}

func TestWriteFeatureFiles_WritesSanitizedFeatureFile(t *testing.T) {
	docs, err := parseSpecSources("../core/testdata/sample.spec.ts")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}

	targetDir := t.TempDir()
	if err := writeFeatureFiles(docs, targetDir); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	uri := core.CalculateFeaturePath(docs[0].Uri)
	relativeDir, baseName := featureFilePathPartsFromURI(uri)
	featurePath := filepath.Join(targetDir, filepath.FromSlash(relativeDir), baseName+".feature")
	content, err := os.ReadFile(featurePath)
	if err != nil {
		t.Fatalf("expected feature file to be written at %s: %v", featurePath, err)
	}

	text := string(content)
	if !strings.Contains(text, "Feature: User authentication") {
		t.Fatalf("feature output missing feature title: %q", text)
	}
	if !strings.Contains(text, "Scenario: log in successfully") {
		t.Fatalf("feature output missing expected scenario: %q", text)
	}
	if !strings.Contains(text, "Given the user is on the login page") {
		t.Fatalf("feature output missing expected step: %q", text)
	}
}
