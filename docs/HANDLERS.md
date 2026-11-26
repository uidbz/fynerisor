# Scene handlers

A scene page can stay declarative. When it needs behaviour, the sidecar
`page.risor` follows one small convention, so the designer can take you from a
widget event to a standalone function:

- one function per node event, named `on_<node>_<event>`;
- handlers are ordinary `let` functions, not inline functions in `page.Bind`;
- a `page.Bind({...})` section, usually at the bottom of the file, lists the
  handler names.

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

Different node/event pairs can map to the same name (`table-cell`/`tapped`
and `table`/`cellTapped`, or `a-b` and `a--b`). When the rewrite or the
designer generates a name that another node event already uses, it adds a
suffix (`on_table_cell_tapped_2`) and reports a warning.

A handler name must be a Risor identifier to be a `let` function and a
shorthand binding. Validation warns about other names (`my-handler`); they
work only with a quoted key, `page.Bind({"my-handler": fn})`, and the rewrite
leaves them alone.

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

`-json` prints the report (created, converted, bound, warnings) as JSON.

It will:

- give every `{"do":"call"}` action a handler name if it has none;
- move inline `page.Bind` functions to standalone `let` declarations directly
  above their `page.Bind` call, leaving a shorthand entry (`hello,`) in place;
- create a stub for a handler that is neither declared nor bound;
- add names that are not bound yet to the first `page.Bind({...})` map, or
  append a new `page.Bind` section when there is none.

The rewrite edits the script in place. Each `page.Bind` call stays where it
is (calls are neither merged nor moved), comments and formatting are kept,
and a script that needs no change is not rewritten. A lifted function keeps
its body; only its indentation changes (multi-line strings excepted). It
leaves these alone and reports a warning where it matters:

- an inline function whose name is already declared (a `let`, `const`,
  `function` or an earlier lifted handler) stays inline;
- keys that are not identifiers (`"a-b"`, template keys) and spreads;
- `page.Bind` calls inside a function or block, and calls whose argument is
  not a map literal;
- a script that does not parse: the document still gets its handler names,
  the script is unchanged, so a half-written handler never blocks a save.

Files are replaced atomically and keep their mode; CRLF files stay CRLF.

The CLI designer and the reference browser host run the same pass on save
through `designer.Options.BeforeSave` and
`scene/handlers.ApplySidecar`. Local sidecars are rewritten next to the
document; remote sidecars are left to the host, but empty handler names are
still filled in. The designer records those name changes as an undoable
edit.

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

Current limitations:

- An expression sees only the fields of the values map: no builtins (`len`,
  `string`, ...) and no `{` or `}` characters inside `${...}`.
- The map is re-evaluated after `page.Bind` and `node.On` handlers and after
  `page.BindValues` itself. State changed elsewhere (a timer, a goroutine)
  needs `page.RefreshBindings()`.
- Until `page.BindValues` runs, a property shows its raw `${...}` text.
- Only number, integer and bool properties receive raw values; every other
  kind, including lists, receives the string form.

Keep binding expressions pure: read values, no side effects. That is the
subset that can move from v2 to v3 later without changing the scene file.

## Why this shape

- The handler is easy to find by name: node and event are in the name.
- The bindings section is the single place that shows what is wired.
- The designer can create/open a handler without parsing arbitrary inline
  code.
- The handler file reads top to bottom: state, handlers, bindings.

The designer's "+ handler" and "Edit handler" buttons use this convention:
they name the call action and ask the host to open the handler
(`designer.Options.EditHandler`). Handlers stay on Risor v2; the optional
binding expressions are the part that may move to v3 later.
