# Issue: Silent Failures on Missing Widget Methods

## Problem

When a script calls a method that doesn't exist on a widget (e.g., `progressBar.Hide()`), the script fails silently without showing a clear error message to the user.

## Example

```risor
let progressBar = widget.NewProgressBarInfinite()
progressBar.Stop()
progressBar.Hide()  // This method doesn't exist - causes silent failure
```

**Expected**: Clear error dialog: "Method 'Hide' not found on ProgressBarInfinite widget"

**Actual**: Script silently fails, window shows default "Something went wrong!" content, no error visible to user

## Root Cause

The error IS being caught and passed to `SetStatus("ERROR: " + err.Error())` in `window.go:299`, but:
1. The error message from Risor might not be descriptive enough
2. The consuming application (goto) might not be displaying the error properly
3. There's no fallback error UI if the consuming app doesn't handle it

## Current Behavior

```go
// In window.go Execute()
result, err := w.runner.Eval()
if err != nil {
    if ctx.Err() == context.Canceled {
        return
    }
    w.SetStatus("ERROR: " + err.Error())  // Error is reported here
    return
}
```

The error is reported via `SetStatus()`, but:
- If no status callback is set, the error is lost
- If the consuming app's status bar isn't visible, user doesn't see it
- The window still shows default content instead of a clear error UI

## Proposed Solutions

### Option 1: Built-in Error Dialog
When script execution fails, show a fynerisor-provided error dialog:

```go
if err != nil {
    if ctx.Err() == context.Canceled {
        return
    }
    w.SetStatus("ERROR: " + err.Error())
    
    // Show built-in error dialog
    w.showScriptError("Script Execution Error", err.Error())
    return
}
```

### Option 2: Better Error Content
Replace the window content with an error display:

```go
if err != nil {
    w.SetStatus("ERROR: " + err.Error())
    
    // Show error in window content
    errorLabel := widget.NewLabel("Script Error:\n\n" + err.Error())
    errorLabel.Wrapping = fyne.TextWrapWord
    w.FyneWindow.SetContent(container.NewCenter(errorLabel))
    return
}
```

### Option 3: Logging
At minimum, log to stderr so errors are visible in terminal:

```go
if err != nil {
    if ctx.Err() == context.Canceled {
        return
    }
    log.Printf("Script execution error: %v", err)
    w.SetStatus("ERROR: " + err.Error())
    return
}
```

## Workaround for Now

In consuming applications, implement proper error handling in the `SetStatus` callback:

```go
gui.WithStatusCallback(func(status string) {
    if strings.HasPrefix(status, "ERROR: ") {
        // Show error dialog or page
        showErrorDialog(status)
    }
})
```

## Related

This would also help catch:
- Undefined variables
- Type errors
- Missing required modules
- Syntax errors that slip past analysis

All errors should be visible to users, not silently fail.
