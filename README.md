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
go run main.go rev --help
```

## Key features

- Reverse engineer feature files from test files (Jasmine, Go, Java)
- Create test file scaffold from feature files (Angular, Go, Java)
- Export feature files into Jira syntax

## References

- https://cucumber.io/docs/gherkin/reference/
- https://github.com/cucumber/gherkin
- https://marketplace.visualstudio.com/items?itemName=alexkrechik.cucumberautocomplete
