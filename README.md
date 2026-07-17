# Gherkin

This is a commandline utility tool that can generate "feature" files in Gherkin syntax from the tests of source code and vice versa.

The tool is not generally applicable to all projects in the world, but makes the assumption that projects that use the "gherkin" tool write the tests having the Gherkin style in mind:

- Every test file represents on Feature

Like the Dart programming language has been designed to contain only features that can be compiled to JavaScript, the Gherkin tool can only process source code that uses framework features that can be conceptually mapped to Gherkin features. The tool rejects to process files that do not abide by these assumptions.

For example, in TypeScript projects:

- "describe" block is mapped to "Feature"
- "beforeEach" is mapped to "Background"
- Nested "describe" block is mapped to "Rule"
- "it" is mapped to "Scenario"

Key features

- Reverse engineer feature files from test files (Jasmine, Go, Java)
- Create test file scaffold from feature files (Angular, Go, Java)
- Export feature files into Jira syntax

## Installation

```sh
go install
```

## Development

### Run tests

```sh
go test ./...
```

## Integration tests

```sh
TEMP_DIR=$(mktemp -d)
go run main.go rev --source "$SOURCE_DIR" --target "$TEMP_DIR" && code $TEMP_DIR
```

## Maintenance

### Open GitHub issue in the browser

If the GitHub CLI tool `gh` is installed, the description for a ticket can be opened like this:

```sh
ISSUE_NUMBER=10
gh issue view ${ISSUE_NUMBER} --json url
```

## Alternatives

- [Cucumber](https://cucumber.io/)
- [Process Extraction from Text](https://arxiv.org/pdf/2110.03754)
- Annotated Textual Descriptions of Processes (ATDP)

## References

- https://cucumber.io/docs/gherkin/reference/
- https://github.com/cucumber/gherkin
- https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete
