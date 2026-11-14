package v3scene

// The portable type text. Named types are declared in this order, because a
// named type may only use opaque types and named types declared before it.
const (
	typeCustomer = `{name: string, city: string, active: bool}`

	typeState = `{count: int, name: string, city: string, customers: Customer[], log: string[]}`

	// typeEvent is open: the renderer attaches component payload fields
	// (row, col, text, item, ...) and we do not want the checker to reject
	// them. Optional fields keep init and keyboard events cheap.
	typeEvent = `{type: string, node?: string, values?: {name?: string, city?: string, ...}, data?: Customer[], ...}`

	// typeAction is a tagged union: "do" is the tag. Each arm is an exact
	// record, so a misspelled field or a wrong value kind is a load error.
	typeAction = `{do: "set", node: string, prop: string, value: any}` +
		` | {do: "show", node: string}` +
		` | {do: "hide", node: string}` +
		` | {do: "toggle", node: string}` +
		` | {do: "rows", node: string, rows: string[][]}` +
		` | {do: "append", node: string, text: string}` +
		` | {do: "status", text: string}` +
		` | {do: "fetch", node: string, target: string}` +
		` | {do: "save", customer: Customer}`

	typeResult = `{state: State, actions: Action[]}`
)

// Customer is a row of the demo's tiny data model. It is a Go struct so v3
// derives a precise record type from it (TypeOf) and the typed layer can
// read its fields without reflection.
type Customer struct {
	Name   string `json:"name"`
	City   string `json:"city"`
	Active bool   `json:"active"`
}

// State is owned by the host. The program receives a value of this type and
// returns the next one; it is never mutated in place.
type State struct {
	Count     int        `json:"count"`
	Name      string     `json:"name"`
	City      string     `json:"city"`
	Customers []Customer `json:"customers"`
	Log       []string   `json:"log"`
}

// Result is what every update program returns.
type Result struct {
	State   State    `json:"state"`
	Actions []Action `json:"actions"`
}

// Action is the host-side shape of the tagged union in typeAction. Optional
// fields are pointers or carry omitempty so the v3 decode accepts every arm.
// The union itself is checked in Risor, before the value reaches this struct.
type Action struct {
	Do       string     `json:"do"`
	Node     string     `json:"node,omitempty"`
	Prop     string     `json:"prop,omitempty"`
	Value    any        `json:"value,omitempty"`
	Text     string     `json:"text,omitempty"`
	Rows     [][]string `json:"rows,omitempty"`
	Target   string     `json:"target,omitempty"`
	Customer *Customer  `json:"customer,omitempty"`
}
