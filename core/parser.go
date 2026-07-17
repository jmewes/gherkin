package core

import (
	"os"
	"regexp"
	"strings"

	messages "github.com/cucumber/messages/go/v28"
	"github.com/experimental-software/gherkin/utils"
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
	kind       string
	name       string   // describe or it title
	ancestors  []string // ancestor describe names (for describe frames)
	depth      int      // brace depth when this frame was opened
	steps      []*messages.Step
	scenarios  []*messages.Scenario
	background []*messages.Step
	children   []*frame // nested describe frames
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
	var topLevelDescribes []*frame
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
				if len(top.ancestors) > 0 {
					// Attach to enclosing describe for later processing.
					for i := len(stack) - 1; i >= 0; i-- {
						if stack[i].kind == frameDescribe {
							stack[i].children = append(stack[i].children, top)
							break
						}
					}
				} else {
					topLevelDescribes = append(topLevelDescribes, top)
				}
			}
		}
	}

	var docs []*messages.GherkinDocument
	for _, d := range topLevelDescribes {
		emitDescribe(d, path, nil, &docs)
	}
	return docs, nil
}

// hasTests reports whether the describe frame contains scenarios directly or
// in any transitively nested describe.
func hasTests(d *frame) bool {
	if len(d.scenarios) > 0 {
		return true
	}
	for _, c := range d.children {
		if hasTests(c) {
			return true
		}
	}
	return false
}

// emitDescribe converts a describe frame (and its nested describes) into one
// or more GherkinDocuments and appends them to docs. The uriAncestors argument
// contains the describe names that lead to the current frame but were skipped
// during emission (i.e. their parent describes had no scenarios and were not
// themselves emitted as Features).
func emitDescribe(d *frame, path string, uriAncestors []string, docs *[]*messages.GherkinDocument) {
	var testChildren []*frame
	for _, c := range d.children {
		if hasTests(c) {
			testChildren = append(testChildren, c)
		}
	}

	// Case: describe with no test-bearing children.
	if len(testChildren) == 0 {
		if len(d.scenarios) == 0 {
			return
		}
		*docs = append(*docs, buildFeatureDoc(d, path, uriAncestors))
		return
	}

	// Case: describe has scenarios and nested test-bearing describes.
	// Emit this describe as its own Feature (with just its direct
	// scenarios), and each test-bearing child as a separate top-level
	// Feature.
	if len(d.scenarios) > 0 {
		*docs = append(*docs, buildFeatureDoc(d, path, uriAncestors))
		childAncestors := append(append([]string{}, uriAncestors...), d.name)
		for _, c := range testChildren {
			emitDescribe(c, path, childAncestors, docs)
		}
		return
	}

	childAncestors := append(append([]string{}, uriAncestors...), d.name)
	for _, c := range testChildren {
		emitDescribe(c, path, childAncestors, docs)
	}
}

func buildFeatureDoc(d *frame, path string, uriAncestors []string) *messages.GherkinDocument {
	var children []*messages.FeatureChild
	if len(d.background) > 0 {
		children = append(children, &messages.FeatureChild{
			Background: &messages.Background{Steps: d.background},
		})
	}
	for _, sc := range d.scenarios {
		children = append(children, &messages.FeatureChild{Scenario: sc})
	}

	var uriParts []string
	for _, p := range append(append([]string{}, uriAncestors...), d.name) {
		uriParts = append(uriParts, utils.ToKebabCase(p))
	}
	uri := path + "/" + strings.Join(uriParts, "/")
	if idx := strings.Index(uri, "src/app/"); idx >= 0 {
		uri = uri[idx+len("src/app/"):]
	}

	return &messages.GherkinDocument{
		Uri: uri,
		Feature: &messages.Feature{
			Name:     d.name,
			Children: children,
		},
	}
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
