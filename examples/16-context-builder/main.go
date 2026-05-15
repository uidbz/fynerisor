package main

import (
	"fmt"
	"log"
	"os"

	"github.com/uidbz/fynerisor"
)

func main() {
	// Example 1: Simple eval without imports
	ctx := fynerisor.NewContext(
		fynerisor.WithHTTP(),
		fynerisor.WithOS(),
	)

	script1 := `
		require(["v0.2", "@http", "@os"])

		let platform = os.goos()
		print("Running on:", platform)

		// Simple calculation
		let numbers = [1, 2, 3, 4, 5].map(x => x * 2)
		let sum = 0
		numbers.each(n => { sum = sum + n })
		print("Sum of doubled numbers:", sum)

		sum
	`

	fmt.Println("=== Example 1: Simple eval ===")
	result, err := ctx.Eval(script1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Result:", result)

	// Example 2: Eval with imports
	fmt.Println("\n=== Example 2: With imports ===")

	// Create a simple utils.risor file
	utilsScript := `
		let double = (x) => x * 2
		let triple = (x) => x * 3
		let greet = (name) => "Hello, " + name + "!"
	`

	err = os.WriteFile("/tmp/test-utils.risor", []byte(utilsScript), 0644)
	if err != nil {
		log.Fatal(err)
	}

	mainScript := `
		import("/tmp/test-utils.risor")
		require("v0.2")

		let result1 = double(5)
		let result2 = triple(4)
		let greeting = greet("World")

		print("double(5) =", result1)
		print("triple(4) =", result2)
		print(greeting)

		[result1, result2, greeting]
	`

	fetchFunc := func(path string) (string, error) {
		data, err := os.ReadFile(path)
		return string(data), err
	}

	result, err = ctx.EvalWithImports(mainScript, fetchFunc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Result:", result)

	fmt.Println("\n=== All examples completed successfully ===")
}
