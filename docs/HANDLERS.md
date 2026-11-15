# Scene handlers

A scene page can stay declarative. When it needs behaviour, the sidecar
`page.risor` follows one small convention, so the designer can take you from a
widget event to a standalone function:

- one function per node event, named `on_<node>_<event>`;
- handlers are ordinary `let` functions, not inline functions in `page.Bind`;
- one `page.Bind({...})` section at the bottom of the file.

```risor
require(["v0.10", "@gui", "@scene"])

let count = 0

let on_inc_tapped = (evt) => {
    count = count + 1
}

page.Bind({
    on_inc_tapped,
})
```

The scene document points at the handler with a normal `call` action:

```json
"on": {
  "tapped": [{"do": "call", "handler": "on_inc_tapped"}]
}
```

Node IDs use lowercase letters, digits and dashes. A dash becomes `_` in the
handler name, and camelCase event names become snake_case:

| Node | Event | Handler |
|---|---|---|
| `submit-btn` | `tapped` | `on_submit_btn_tapped` |
| `name-entry` | `changed` | `on_name_entry_changed` |
| `table` | `cellTapped` | `on_table_cell_tapped` |

## Converting an existing page

Older pages often put handlers inline:

```risor
page.Bind({
    hello: (evt) => {
        page.Node("title").SetText("Hello")
    }
})
```

Run the rewrite pass:

```sh
fynerisor handlers page.scene.json            # show the result
fynerisor handlers -write page.scene.json     # write script and document
```

It will:

- give every `{"do":"call"}` action a handler name if it has none;
- create a stub for a handler that does not exist;
- move inline `page.Bind` functions to standalone `let` declarations;
- write one bindings section at the bottom.

Existing function bodies are not changed. A handler that is already declared
is only added to the bindings section.

## Bindings

Handlers change state. Bindings show it without a `SetText` call in every
handler:

```risor
let count = 0

let on_inc_tapped = (evt) => {
    count = count + 1
}

page.Bind({
    on_inc_tapped,
})

page.BindValues(() => ({
    count,
}))
```

In the scene document, use the value name in a property:

```json
"props": {"text": "${count}"}
```

`page.BindValues` takes a closure returning a map. The page re-evaluates the
map after every handler and updates every property containing `${name}`.
`${name}` can also be an expression over the map fields, such as
`${count + 1}`. A property that is exactly one placeholder receives the raw
value for number/bool props; text props receive its string form.
`page.RefreshBindings()` forces an update.

Keep binding expressions pure: read values, no side effects. That is the
subset that can move from v2 to v3 later without changing the scene file.

## Why this shape

- The handler is easy to find by name: node and event are in the name.
- The bindings section is the single place that shows what is wired.
- The designer can create/open a handler without parsing arbitrary inline
  code.
- The handler file reads top to bottom: state, handlers, bindings.

This convention is what the planned "Edit handler" button in the designer
will use. Handlers stay on Risor v2; the optional binding expressions are the
part that may move to v3 later.
