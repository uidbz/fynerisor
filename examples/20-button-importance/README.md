# Button Importance Example

Demonstrates button importance levels and styling for visual hierarchy and semantic meaning.

## Features

- **Standard Importance**: High, Medium, Low for visual hierarchy
- **Semantic Importance**: Success (green), Warning (yellow), Danger (red) for meaning
- **Disabled State**: Enable/disable buttons programmatically
- **Constants**: Use `constants.*` for importance values

## Importance Levels

**Standard:**
- `ImportanceHigh` - Prominent, primary actions
- `ImportanceMedium` - Default button appearance
- `ImportanceLow` - Subtle, secondary actions

**Semantic:**
- `SuccessImportance` - Positive/confirmation actions (green)
- `WarningImportance` - Caution/warning actions (yellow)
- `DangerImportance` - Destructive/dangerous actions (red)

## Run

```bash
go run main.go
```

Or with fynerisor CLI:

```bash
fynerisor script.risor
```
