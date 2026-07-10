package core

import (
	"testing"
)

func TestParseFeatureFile(t *testing.T) {
	doc, err := ParseFeatureFile("testdata/sample.feature")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.Feature == nil {
		t.Fatal("expected a feature, got nil")
	}
	expected := "Sample feature"
	if doc.Feature.Name != expected {
		t.Errorf("expected title %q, got %q", expected, doc.Feature.Name)
	}
}

func TestParseFeatureFile_MissingFile(t *testing.T) {
	_, err := ParseFeatureFile("testdata/nonexistent.feature")
	if err == nil {
		t.Fatal("expected an error for missing file, got nil")
	}
}
