# Contributing to Fynerisor

Thank you for your interest in contributing to Fynerisor! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, collaborative, and constructive. We want to create a welcoming environment for everyone.

## How to Contribute

### Reporting Issues

- Check if the issue already exists before creating a new one
- Provide a clear description and reproduction steps
- Include your environment details (OS, Go version, Fyne version)
- If possible, provide a minimal Risor script that demonstrates the issue

### Suggesting Features

- Open an issue describing the feature and its use case
- Explain why it would be valuable to the project
- Be open to discussion about implementation approaches

### Contributing Code

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/bug-description
   ```
4. **Make your changes** (see guidelines below)
5. **Test your changes** thoroughly
6. **Commit your changes** with clear messages
7. **Push to your fork**
8. **Open a Pull Request** on GitHub

## Development Setup

### Prerequisites

- Go 1.25 or later
- Fyne v2.7+ dependencies (see [Fyne documentation](https://docs.fyne.io/started/))

### Building

```bash
# Clone the repository
git clone https://github.com/uidbz/fynerisor.git
cd fynerisor

# Install dependencies
go mod download

# Build the CLI tool
go build ./cmd/fynerisor

# Run tests
go test ./...
```

### Running Examples

```bash
cd examples/01-hello-world
go run main.go
# or use the CLI
../../fynerisor script.risor
```

## Code Style Guidelines

### General Principles

- Write clean, readable code
- Follow standard Go conventions and idioms
- Keep functions focused and reasonably sized
- Add comments for non-obvious code
- Write tests for new functionality

### Go Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` to format code (runs automatically with most editors)
- Keep line length reasonable (120 characters is a soft limit)
- Use meaningful variable names

### Risor Script Style

- Use Risor v2 syntax (arrow functions or `function` keyword, no C-style loops)
- Prefer functional iteration (`.each()`, `.map()`, `.filter()`)
- Use clear, descriptive variable names
- Add comments to explain complex logic

### Project Structure

```
fynerisor/
├── *.go              # Core library files (app, window, context, etc.)
├── modules/          # Risor modules (http, sql, os, time, etc.)
│   ├── http/
│   ├── sql/
│   └── ...
├── widget/           # Widget implementations
├── container/        # Container implementations  
├── canvas/           # Canvas object implementations
├── binding/          # Data binding implementations
├── cmd/fynerisor/    # CLI tool
├── examples/         # Example scripts
└── docs/            # Documentation
```

### Adding New Widgets

When adding support for a new Fyne widget:

1. Add the widget method to `widget/widget.go` or create a new file in `widget/`
2. Use the IIFE pattern for callbacks to avoid closure capture bugs:
   ```go
   button := func(f *object.Closure) object.Object {
       return risorwidget.NewButton(name, func() {
           obj.w.Do(func() {
               _, err := callFunc(ctx, f, []object.Object{})
               // handle error
           })
       })
   }(fn)
   ```
3. Add documentation in `llms.txt` with usage examples
4. Create an example in `examples/`
5. Update `docs/WIDGET_STATUS.md`

### Adding New Modules

When adding a new Risor module:

1. Create a new directory in `modules/`
2. Implement the module following the existing module patterns (see `modules/http/` or `modules/os/`)
3. Add a `WithModuleName()` option in `options.go`
4. Add documentation in `llms.txt`
5. Create an example demonstrating the module
6. Add tests if possible

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./widget

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestComplexLayout
```

### Writing Tests

- Add tests for new functionality
- Test both success and error cases
- Use `fyne.io/fyne/v2/test` for GUI testing
- Keep tests focused and independent

Example test structure:
```go
func TestNewWidget(t *testing.T) {
    a := test.NewApp()
    defer a.Quit()
    w := a.NewWindow("Test")
    fw := NewWindow(w)

    script := `
    let widget = widget.NewSomething("test")
    window.SetContent(widget)
    `

    fw.LoadScript(script)
    fw.Execute()
    time.Sleep(200 * time.Millisecond)

    if fw.Status != "Ready!" {
        t.Errorf("Expected Ready!, got %s", fw.Status)
    }
}
```

## Commit Messages

Write clear, descriptive commit messages:

```
Brief summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain what changed and why, not how (the diff shows how).

- Bullet points are fine
- Reference issues: Fixes #123
```

Examples:
- `Add support for CheckGroup widget`
- `Fix closure capture bug in button callbacks`
- `Update HTTP module to support POST with auth`
- `Docs: Add example for data binding`

## Pull Request Process

1. **Update documentation** if you're changing functionality
2. **Add examples** for new features
3. **Ensure all tests pass** (`go test ./...`)
4. **Update CHANGELOG.md** if applicable
5. **Keep PRs focused** - one feature/fix per PR when possible
6. **Respond to review feedback** constructively

### PR Title Format

- `Feature: Add support for XYZ widget`
- `Fix: Resolve closure capture in callbacks`
- `Docs: Improve installation instructions`
- `Refactor: Reorganize module structure`
- `Test: Add tests for data binding`

## Areas for Contribution

We especially welcome contributions in these areas:

### High Priority
- **Widget coverage** - Add support for remaining Fyne widgets
- **Test coverage** - Add tests for existing functionality
- **Documentation** - Improve docs, add more examples
- **Bug fixes** - Fix reported issues

### Medium Priority  
- **Performance** - Optimize hot paths
- **Error messages** - Improve error reporting and messages
- **Module features** - Extend existing modules with new functions

### Nice to Have
- **CI/CD** - Set up GitHub Actions for testing
- **Benchmarks** - Add performance benchmarks
- **Platform testing** - Test on Windows/macOS/Linux
- **VSCode extension** - Syntax highlighting for .risor files

## Questions?

- **Issues**: For bugs and feature requests
- **Discussions**: For questions and general discussion
- **Email**: j@uid.bz for private inquiries

## License

By contributing, you agree that your contributions will be licensed under the BSD-3-Clause License.

## Acknowledgments

Contributors are listed in the project's commit history. Thank you to everyone who helps improve Fynerisor!

---

**Note**: This is an AI-assisted project (see README for details). Don't let that discourage you from contributing - good code is good code, regardless of how it's written! 🚀
