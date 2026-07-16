package core

import (
	"os"

	gherkin "github.com/cucumber/gherkin/go/v33"
	messages "github.com/cucumber/messages/go/v28"
)

// ParseFeatureFile opens the file at path and parses it into a GherkinDocument.
func ParseFeatureFile(path string) (*messages.GherkinDocument, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := gherkin.ParseGherkinDocument(f, (&messages.UUID{}).NewId)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// FeatureTitle returns the feature title from the document, or an empty string if there is no feature.
func FeatureTitle(doc *messages.GherkinDocument) string {
	if doc.Feature == nil {
		return ""
	}
	return doc.Feature.Name
}
