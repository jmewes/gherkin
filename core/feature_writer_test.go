package core

import (
	"testing"
)

func TestCalculateFeaturePath(t *testing.T) {
	t.Run("nested describe section", func(t *testing.T) {
		// Given a URI for a nested describe section in a test file
		originalUri := "components/input/add-item-button/add-item-button.component.spec.ts/AddItemButtonComponent/NestedSection"

		// When the feature file path is calculated
		actual := CalculateFeaturePath(originalUri)

		// Then for everything after the test file, the "/" is replaced with "."
		expected := "components/input/add-item-button/AddItemButtonComponent.NestedSection"
		if actual != expected {
			t.Errorf("Expected %s, but got %s", expected, actual)
		}
	})

	t.Run("single describe section", func(t *testing.T) {
		// Given a URI for a single describe section in a test file
		originalUri := "components/input/add-item-button/add-item-button.component.spec.ts/AddItemButtonComponent"

		// When the feature file path is calculated
		actual := CalculateFeaturePath(originalUri)

		// Then the test file name gets discarded
		expected := "components/input/add-item-button/AddItemButtonComponent"
		if actual != expected {
			t.Errorf("Expected %s, but got %s", expected, actual)
		}
	})
}
