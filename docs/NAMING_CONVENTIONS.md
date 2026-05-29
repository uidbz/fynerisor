# Naming Conventions for Fynerisor

This document defines the standard naming conventions used throughout fynerisor to ensure consistency and clarity.

## Module Functions: `snake_case`

All functions exposed through fynerisor modules use **snake_case** naming:

```risor
// OS module
os.current_user()
os.open_browser(url)

// Strings module
strings.replace_all(str, old, new)
strings.to_lower(str)
strings.has_prefix(str, prefix)

// HTTP module
http.get(url)
http.post(url, headers, body)

// Filepath module
filepath.join(parts...)
filepath.base(path)

// Time module
time.now()
time.parse(str)

// IO module
io.cp(src, dst)
io.read_all(path)

// SQL module
conn.query(sql)
conn.exec(sql, args...)
```

**Rationale:**
- Consistency across all modules
- Script-language feel (similar to Python, Ruby, Lua)
- Clear distinction from widget methods
- Easier for users who learn one module function to predict others

## Widget Methods: `PascalCase`

Widget object methods follow Go/Fyne naming conventions using **PascalCase**:

```risor
// Widget creation (factory functions)
widget.NewButton(text, callback)
widget.NewEntry()
widget.NewLabel(text)
widget.NewProgressBar()

// Widget methods
button.SetText("New Text")
button.Disable()
button.Enable()

entry.SetPlaceHolder("Enter text...")
entry.OnChanged(callback)
entry.OnSubmitted(callback)

label.SetText("New Label")

slider.OnChanged(callback)
```

**Rationale:**
- Direct wrappers around Fyne widgets
- Maintains familiarity with underlying Fyne API
- Clear distinction: "this is a widget method, not a module function"
- Go developers working with both Fyne and fynerisor find it natural

## Container Functions: `PascalCase`

Container creation follows the same pattern as widgets:

```risor
container.NewVBox(widgets...)
container.NewHBox(widgets...)
container.NewBorder(top, bottom, left, right, center)
container.NewScroll(content)
```

## Data Binding: `PascalCase`

Binding functions follow Go conventions:

```risor
binding.NewString(value)
binding.NewInt(value)
binding.NewFloat(value)
binding.NewBool(value)
```

## Summary

| Category | Convention | Example |
|----------|-----------|---------|
| Module functions | `snake_case` | `io.read_all()`, `strings.to_lower()` |
| Widget methods | `PascalCase` | `button.SetText()`, `entry.Disable()` |
| Widget factories | `PascalCase` | `widget.NewButton()`, `widget.NewEntry()` |
| Container factories | `PascalCase` | `container.NewVBox()` |
| Binding factories | `PascalCase` | `binding.NewString()` |
| Properties (read/write) | `PascalCase` | `button.Text`, `entry.PlaceHolder` |

## When Adding New Features

1. **New module function?** Use `snake_case`
2. **New widget/object method?** Use `PascalCase`
3. **New factory function?** Use `PascalCase` with `New` prefix
4. **Not sure?** Follow the pattern of similar existing functions

This convention ensures a consistent and predictable API for users while honoring both the Risor scripting language and the underlying Fyne framework.
