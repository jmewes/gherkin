package core

import "strings"

func CalculateFeaturePath(originalUri string) string {
	parts := strings.Split(originalUri, "/")
	var beforeParts []string
	var afterParts []string
	foundSpec := false
	for _, part := range parts {
		if strings.HasSuffix(part, ".spec.ts") {
			foundSpec = true
			continue
		}
		if foundSpec {
			afterParts = append(afterParts, part)
		} else {
			beforeParts = append(beforeParts, part)
		}
	}
	result := strings.Join(beforeParts, "/")
	if len(afterParts) > 0 {
		if result != "" {
			result += "/"
		}
		result += strings.Join(afterParts, ".")
	}
	return result
}
