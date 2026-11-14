// Package v3scene is a prototype of a Risor v3 scene-page runtime.
//
// It implements the model agreed in deepnoodle-ai/risor#641: the host owns
// the state and the event loop, runs one immutable v3 program per event, and
// applies the actions the program returns. There are no script callbacks and
// no VM shared with the UI thread.
//
// Two effects models are provided, so they can be compared on the same page:
//
//   - ModeActions: side effects are actions. The program returns
//     {do: "fetch"} or {do: "save"}, and the host performs them.
//   - ModeHostFunctions: the program calls synchronous host functions such as
//     store.customers() and store.add() during the run.
//
// The classic fynerisor widget API (v2, closures) is untouched.
package v3scene
