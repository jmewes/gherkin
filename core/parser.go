package core

import (
	"errors"
	"os"
	"regexp"
	"strings"

	messages "github.com/cucumber/messages/go/v28"
)

// normalizeScenarioTitle strips a leading "should " prefix (case-insensitive).
func normalizeScenarioTitle(title string) string {
	if strings.HasPrefix(strings.ToLower(title), "should ") {
		return title[len("should "):]
	}
	return title
}

var stepCommentRe = regexp.MustCompile(`(?i)^\s*//\s*(given|when|then|and|but)\s+(.+)$`)

// matchStepComment returns the keyword (e.g. "Given ") and the remainder text
// if line matches a Gherkin step comment, otherwise ok is false.
func matchStepComment(line string) (keyword, text string, ok bool) {
	m := stepCommentRe.FindStringSubmatch(line)
	if m == nil {
		return "", "", false
	}
	kw := strings.ToUpper(m[1][:1]) + strings.ToLower(m[1][1:])
	return kw + " ", strings.TrimSpace(m[2]), true
}

// extractQuotedName returns the first single- or double-quoted string in s.
func extractQuotedName(s string) (string, bool) {
	for _, q := range []byte{'\'', '"'} {
		start := strings.IndexByte(s, q)
		if start < 0 {
			continue
		}
		end := strings.IndexByte(s[start+1:], q)
		if end < 0 {
			continue
		}
		return s[start+1 : start+1+end], true
	}
	return "", false
}

const (
	frameDescribe   = "describe"
	frameIt         = "it"
	frameBeforeEach = "beforeEach"
)

type frame struct {
	kind                string
	name                string   // describe or it title
	ancestors           []string // ancestor describe names (for describe frames)
	depth               int      // brace depth when this frame was opened
	steps               []*messages.Step
	scenarios           []*messages.Scenario
	background          []*messages.Step
	rules               []*messages.Rule // nested describe blocks interpreted as Rules
	nestedDescribeCount int              // number of nested describe blocks opened directly inside this describe
}

// ParseSpecFile reads a Jasmine *.spec.ts file and converts its
// describe/it/beforeEach structure into GherkinDocuments.
func ParseSpecFile(path string) ([]*messages.GherkinDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	var stack []*frame
	var docs []*messages.GherkinDocument
	braceDepth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect block openers before counting braces on this line.
		var opened *frame
		if strings.Contains(trimmed, "describe(") {
			if name, ok := extractQuotedName(trimmed[strings.Index(trimmed, "describe(")+len("describe("):]); ok {
				// ancestor names = all current describe frame names
				var ancestors []string
				for _, f := range stack {
					if f.kind == frameDescribe {
						ancestors = append(ancestors, f.name)
					}
				}
				opened = &frame{kind: frameDescribe, name: name, ancestors: ancestors}
				// At most one level of nesting is supported: a single nested
				// describe is interpreted as a Rule. Deeper nesting is rejected.
				if len(ancestors) > 1 {
					return nil, errors.New("nested describe blocks are not supported when they contain tests")
				}
				// A Feature may contain at most one Rule. Multiple nested
				// describe blocks inside the same describe are rejected.
				if len(ancestors) == 1 {
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].kind == frameDescribe {
							stack[i].nestedDescribeCount++
							if stack[i].nestedDescribeCount > 1 {
								return nil, errors.New("nested describe blocks are not supported when they contain tests")
							}
							break
						}
					}
				}
			}
		} else if strings.Contains(trimmed, "beforeEach(") {
			opened = &frame{kind: frameBeforeEach}
		} else if strings.Contains(trimmed, "it(") {
			// avoid matching "fit(" or "xit(" etc. by checking the char before "it("
			idx := strings.Index(trimmed, "it(")
			if idx == 0 || !isIdentChar(trimmed[idx-1]) {
				if name, ok := extractQuotedName(trimmed[idx+len("it("):]); ok {
					opened = &frame{kind: frameIt, name: name}
				}
			}
		}

		// Count braces on this line.
		for _, ch := range trimmed {
			if ch == '{' {
				braceDepth++
			} else if ch == '}' {
				braceDepth--
			}
		}

		// If we opened a new frame, record its depth after counting (so we
		// know when the matching closing brace brings us back here).
		if opened != nil {
			opened.depth = braceDepth
			stack = append(stack, opened)
		}

		// Collect step comments for it/beforeEach frames.
		if len(stack) > 0 {
			top := stack[len(stack)-1]
			if top.kind == frameIt || top.kind == frameBeforeEach {
				if kw, txt, ok := matchStepComment(line); ok {
					top.steps = append(top.steps, &messages.Step{Keyword: kw, Text: txt})
				}
			}
		}

		// Check whether the top frame has closed (brace depth returned to its open depth - 1).
		for len(stack) > 0 {
			top := stack[len(stack)-1]
			if braceDepth >= top.depth {
				break
			}
			// Pop frame.
			stack = stack[:len(stack)-1]

			switch top.kind {
			case frameIt:
				if len(top.steps) > 0 {
					scenario := &messages.Scenario{
						Name:  normalizeScenarioTitle(top.name),
						Steps: top.steps,
					}
					// Attach to nearest enclosing describe.
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].kind == frameDescribe {
							stack[i].scenarios = append(stack[i].scenarios, scenario)
							break
						}
					}
				}

			case frameBeforeEach:
				var bgSteps []*messages.Step
				for _, s := range top.steps {
					kw := strings.TrimRight(s.Keyword, " ")
					if strings.EqualFold(kw, "given") || strings.EqualFold(kw, "and") {
						bgSteps = append(bgSteps, s)
					}
				}
				if len(bgSteps) > 0 {
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].kind == frameDescribe {
							stack[i].background = bgSteps
							break
						}
					}
				}

			case frameDescribe:
				// A nested describe is interpreted as a Rule and attached to
				// its enclosing Feature describe.
				if len(top.ancestors) > 0 {
					if len(top.scenarios) == 0 {
						break
					}
					var ruleChildren []*messages.RuleChild
					if len(top.background) > 0 {
						ruleChildren = append(ruleChildren, &messages.RuleChild{
							Background: &messages.Background{Steps: top.background},
						})
					}
					for _, sc := range top.scenarios {
						ruleChildren = append(ruleChildren, &messages.RuleChild{Scenario: sc})
					}
					rule := &messages.Rule{
						Name:     top.name,
						Children: ruleChildren,
					}
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].kind == frameDescribe {
							stack[i].rules = append(stack[i].rules, rule)
							break
						}
					}
					break
				}

				if len(top.scenarios) == 0 && len(top.rules) == 0 {
					break
				}
				// Build children: background first, then scenarios, then rules.
				var children []*messages.FeatureChild
				if len(top.background) > 0 {
					children = append(children, &messages.FeatureChild{
						Background: &messages.Background{Steps: top.background},
					})
				}
				for _, sc := range top.scenarios {
					children = append(children, &messages.FeatureChild{Scenario: sc})
				}
				for _, r := range top.rules {
					children = append(children, &messages.FeatureChild{Rule: r})
				}
				// Uri = path + "/" + joined ancestor describe names + "/" + this name.
				var uriParts []string
				for _, p := range append(top.ancestors, top.name) {
					uriParts = append(uriParts, sanitizeURI(p))
				}
				uri := path + "/" + strings.Join(uriParts, "/")
				// Extract substring after "src/app/"
				if idx := strings.Index(uri, "src/app/"); idx >= 0 {
					uri = uri[idx+len("src/app/"):]
				}

				doc := &messages.GherkinDocument{
					Uri: uri,
					Feature: &messages.Feature{
						Name:     top.name,
						Children: children,
					},
				}
				docs = append(docs, doc)
			}
		}
	}

	return docs, nil
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func sanitizeURI(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	result := b.String()
	return regexp.MustCompile(`_+`).ReplaceAllString(result, "_")
}
