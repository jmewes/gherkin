package core

import (
	"testing"

	messages "github.com/cucumber/messages/go/v28"
)

func TestMinimalTestFile(t *testing.T) {
	// Given a test file that has a "describe" block
	// And two "it" blocks
	specPath := "testdata/sample.spec.ts"

	// When the Gherkin document is parsed
	docs, err := ParseSpecFile(specPath, false)
	if err != nil {
		t.Fatalf("ParseSpecFile failed: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}

	doc := docs[0]

	// Then the description of the "describe" block becomes the Feature name of the Gherkin document
	if doc.Feature.Name != "User authentication" {
		t.Errorf("Expected Feature name 'User authentication', got '%s'", doc.Feature.Name)
	}

	// And the "it" blocks are mapped to Scenarios
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

func TestParseSpecFile_relaxed_enabled(t *testing.T) {
	// Given a test file with a test without Gherkin steps
	specPath := "testdata/test-without-gherkin-steps.spec.ts"
	// And the relaxed mode is enabled
	option := withRelaxed(true)

	// When the spec file gets parsed
	docs := mustParseSpecFile(specPath, option)

	// Then the test is included
	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}
	doc := docs[0]
	if len(doc.Feature.Children) != 2 {
		t.Fatalf("Expected 2 scenarios, got %d", len(doc.Feature.Children))
	}
}

func TestParseSpecFile_relaxed_disabled(t *testing.T) {
	// Given a test file with a test without Gherkin steps
	specPath := "testdata/test-without-gherkin-steps.spec.ts"
	// And the relaxed mode is disabled
	option := withRelaxed(false)

	// When the spec file gets parsed
	docs := mustParseSpecFile(specPath, option)

	// Then the test is included
	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}
	doc := docs[0]
	if len(doc.Feature.Children) != 1 {
		t.Fatalf("Expected 1 scenario, got %d", len(doc.Feature.Children))
	}
}

func TestResourcePathSimpleFeature(t *testing.T) {
	// Given a test file with a single "describe" block
	// And a single test
	specPath := "testdata/resource_Name_test/one_feature-with-one-test.spec.ts"

	// When the spec file gets parsed
	docs := mustParseSpecFile(specPath)

	// Then the resource path has the path elements separated by slashes,
	// And the path elements themselves should be in lower-kebab-case
	// And ignores the TypeScript specific file suffix
	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}
	doc := docs[0]
	if doc.Uri != "testdata/resource-name-test/one-feature-with-one-test/just-a-test" {
		t.Errorf("Unexpected resource path: '%s'", doc.Uri)
	}
}

func TestResourcePathNestedDescribeBlock(t *testing.T) {
	// Given a test file with a single "describe" block
	// And a single test
	specPath := "testdata/resource_Name_test/with-nested-describe-block.spec.ts"

	// When the spec file gets parsed
	docs := mustParseSpecFile(specPath)

	// Then the resource path has the nested "describe" blocks as separate path elements,
	// And the path elements have CamelCase converted to lower-kebab-case
	if len(docs) != 1 {
		t.Fatalf("Expected 1 GherkinDocument, got %d", len(docs))
	}
	doc := docs[0]
	if doc.Uri != "testdata/resource-name-test/with-nested-describe-block/just-a-test/example-nested-block" {
		t.Errorf("Unexpected resource path: '%s'", doc.Uri)
	}
}

func TestSeparateFeaturesForNestedDescribeBlocks(t *testing.T) {
	// Given a "describe" block without tests (Foo)
	// but a nested "describe" block with tests (Bar),
	// And the nested "describe" block has a nested "describe" block itself (Oogle)
	specPath := "testdata/nested-describe-blocks.spec.ts"

	// When the spec file gets parsed,
	docs := mustParseSpecFile(specPath)

	// Then "Foo" is not interpreted as feature because it has no tests
	if len(docs) != 2 {
		t.Fatalf("Expected 2 GherkinDocuments, got %d", len(docs))
	}
	for _, doc := range docs {
		if doc.Feature.Name == "Foo" {
			t.Errorf("Did not expect a feature named 'Foo'")
		}
	}

	// And "Bar" is interpreted as separate feature with the scenario "should boogle"
	bar := docs[0].Feature
	if bar.Name != "Bar" {
		t.Errorf("Expected first feature name 'Bar', got '%s'", bar.Name)
	}
	var barScenarios []*messages.Scenario
	for _, c := range bar.Children {
		if c.Scenario != nil {
			barScenarios = append(barScenarios, c.Scenario)
		}
	}
	if len(barScenarios) != 1 {
		t.Fatalf("Expected 1 scenario in 'Bar', got %d", len(barScenarios))
	}
	if barScenarios[0].Name != "boogle" {
		t.Errorf("Expected scenario name 'boogle', got '%s'", barScenarios[0].Name)
	}

	// And "Ooogle" is interpreted as separate feature with the scenario "should quux"
	oogle := docs[1].Feature
	if oogle.Name != "Oogle" {
		t.Errorf("Expected second feature name 'Oogle', got '%s'", oogle.Name)
	}
	var oogleScenarios []*messages.Scenario
	for _, c := range oogle.Children {
		if c.Scenario != nil {
			oogleScenarios = append(oogleScenarios, c.Scenario)
		}
	}
	if len(oogleScenarios) != 1 {
		t.Fatalf("Expected 1 scenario in 'Oogle', got %d", len(oogleScenarios))
	}
	if oogleScenarios[0].Name != "quuux" {
		t.Errorf("Expected scenario name 'quuux', got '%s'", oogleScenarios[0].Name)
	}
}

type parseOption func(*parseConfig)

type parseConfig struct {
	relaxed bool
}

func withRelaxed(relaxed bool) parseOption {
	return func(c *parseConfig) {
		c.relaxed = relaxed
	}
}

func mustParseSpecFile(path string, opts ...parseOption) []*messages.GherkinDocument {
	cfg := &parseConfig{relaxed: false}
	for _, opt := range opts {
		opt(cfg)
	}

	docs, err := ParseSpecFile(path, cfg.relaxed)
	if err != nil {
		panic(err)
	}
	return docs
}
