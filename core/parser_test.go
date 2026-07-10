package core

import (
	"testing"
)

func TestMinimalTestFile(t *testing.T) {
	//Given a test file that has a "describe" block
	//And two "it" blocks
	specPath := "testdata/sample.spec.ts"

	//When the Gherkin document is parsed
	docs, err := ParseSpecFile(specPath)
	if err != nil {
		t.Fatalf("ParseSpecFile failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}

	doc := docs[0]

	//Then the description of the "describe" block becomes the Feature name of the Gherkin document
	if doc.Feature.Name != "User authentication" {
		t.Errorf("Expected Feature name 'User authentication', got '%s'", doc.Feature.Name)
	}

	//And the "it" blocks are mapped to Scenarios
	if len(doc.Feature.Children) != 2 {
		t.Fatalf("Expected 2 scenarios, got %d", len(doc.Feature.Children))
	}

	if doc.Feature.Children[0].Scenario.Name != "log in successfully" {
		t.Errorf("Expected first scenario name 'log in successfully', got '%s'", doc.Feature.Children[0].Scenario.Name)
	}

	if doc.Feature.Children[1].Scenario.Name != "handle invalid credentials" {
		t.Errorf("Expected second scenario name 'handle invalid credentials', got '%s'", doc.Feature.Children[1].Scenario.Name)
	}
}

func TestRejectScenariosInNestedDescribe(t *testing.T) {
	//Given a test file where the "it" blocks are in nested "describe" blocks
	specPath := "testdata/invalid-nested-source-file.spec.ts"

	//When the Gherkin document is parsed
	_, err := ParseSpecFile(specPath)

	//Then there is an error
	if err == nil {
		t.Error("Expected an error when parsing a file with nested describe blocks containing it blocks, but got nil")
	}
}

func TestParseTestWithRule(t *testing.T) {
	//Given a test file that has a nested "describe" block
	//And inside the second "describe" block there are "it" blocks
	specPath := "testdata/with-rule.spec.ts"

	//When the Gherkin document is parsed
	docs, err := ParseSpecFile(specPath)
	if err != nil {
		t.Fatalf("ParseSpecFile failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}

	doc := docs[0]

	//Then the first "describe" block is interpreted as "Feature"
	if doc.Feature.Name != "Foo" {
		t.Errorf("Expected Feature name 'Foo', got '%s'", doc.Feature.Name)
	}

	if len(doc.Feature.Children) != 1 {
		t.Fatalf("Expected 1 feature child, got %d", len(doc.Feature.Children))
	}

	//And the second "describe" block is interpreted as "Rule"
	rule := doc.Feature.Children[0].Rule
	if rule == nil {
		t.Fatalf("Expected the feature child to be a Rule, got %+v", doc.Feature.Children[0])
	}
	if rule.Name != "Tax rate of 5.5%" {
		t.Errorf("Expected Rule name 'Tax rate of 5.5%%', got '%s'", rule.Name)
	}

	//And the "it" blocks are interpreted as "Scenarios" inside the "Rule"
	if len(rule.Children) != 1 {
		t.Fatalf("Expected 1 rule child, got %d", len(rule.Children))
	}
	scenario := rule.Children[0].Scenario
	if scenario == nil {
		t.Fatalf("Expected the rule child to be a Scenario, got %+v", rule.Children[0])
	}
	if scenario.Name != "calculate taxes" {
		t.Errorf("Expected scenario name 'calculate taxes', got '%s'", scenario.Name)
	}
}
