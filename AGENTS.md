# Agent guidelines

## Tests

- When provided with an implementation request with Gherkin notation (Given/When/Then), treat it as specification by example.
- In the generated tests, add the steps of the Gherkin notation as comments right before the implementation of the respective step in the test. So there is a comment with a step specification, then some go code, then potentially another comment with a step specification and so on.
- Do not create tests for the "cmd" package.

## Comments

- Each Go source file should have an AGPL license header, using Copyright (C) 2026 John Doe
