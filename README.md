# Sample App with Bad Code

Intentionally poorly written applications in Python and Go, designed as teaching material for code review best practices.

## Structure

```
python-app/   # Python version
  main.py
  pyproject.toml

go-app/       # Go version
  main.go
  go.mod
```

## What's Wrong (and Why It's Useful)

Both apps contain the same categories of issues, translated into each language's idioms (or anti-idioms):

### Security
- Hardcoded API keys, database passwords, and JWT secrets
- MD5 used for password hashing
- SQL injection via string concatenation
- Plaintext password comparison
- Passwords stored in a global list

### Code Quality
- God function (`process` / `Process`) handling unrelated logic via a type string
- Deep nesting with empty else branches
- Vague variable names (`t`, `x`, `d`, `total2`, `flag`, `flag2`)
- Shadowed builtins (Python: `sum`, `max`, `min`)
- Unused imports and variables
- Functions with 15+ positional parameters
- Magic numbers (tax rate, coupon percentages)

### Reliability
- File handles and HTTP response bodies never closed
- Bare/silent exception handling
- 100 retries with no backoff
- Division by zero and index-out-of-range on empty input
- Email sending randomly fails with no logging
- Email validation only checks for `@`
- Password validation accepts any non-empty string

### Design
- No tests, no type hints/structs, no documentation
- No error handling or logging
- Dicts/maps used everywhere instead of proper types
- Malformed connection string construction

### Go-Specific
- `map[string]interface{}` instead of structs
- Errors silently discarded (`_, _` everywhere)
- Deprecated `ioutil.ReadAll`
- `snake_case` instead of `camelCase`
- No `defer` for resource cleanup

## Usage

Use these files as discussion starters in code review workshops. Each function or method contains multiple issues worth identifying and refactoring.

### Python

```bash
cd python-app
python main.py
```

### Go

```bash
cd go-app
go run main.go
```
