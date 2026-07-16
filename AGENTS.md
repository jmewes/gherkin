# Agent guidelines

## Tests

- When provided with an implementation request with Gherkin notation (Given/When/Then), treat it as specification by example.
- In the generated tests, add the steps of the Gherkin notation as comments before the implementation of the respective step in the test.
- Do not create tests for the "cmd" package.

## Comments

- Each Go source file should have an AGPL license header, using Copyright (C) 2026 John Doe
