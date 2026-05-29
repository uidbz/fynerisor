package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deepnoodle-ai/risor/v2"

	"github.com/uidbz/fynerisor/gui"
)

// Sample data structure for the table
type Person struct {
	ID     int
	Name   string
	Email  string
	Status string
}

// Generate sample data
func generatePeople() []Person {
	return []Person{
		{1, "Alice Johnson", "alice@example.com", "Active"},
		{2, "Bob Smith", "bob@example.com", "Active"},
		{3, "Carol Davis", "carol@example.com", "Inactive"},
		{4, "David Wilson", "david@example.com", "Active"},
		{5, "Eve Martinez", "eve@example.com", "Pending"},
		{6, "Frank Brown", "frank@example.com", "Active"},
		{7, "Grace Lee", "grace@example.com", "Active"},
		{8, "Henry Taylor", "henry@example.com", "Inactive"},
		{9, "Iris Anderson", "iris@example.com", "Active"},
		{10, "Jack Thomas", "jack@example.com", "Pending"},
	}
}

func main() {
	// Generate data
	people := generatePeople()

	// Create Risor globals with data access functions
	dataFuncs := map[string]any{
		"data": map[string]any{
			"getCount": func() int {
				return len(people)
			},
			"getPage": func(offset, limit int) [][]string {
				result := [][]string{}
				end := offset + limit
				if end > len(people) {
					end = len(people)
				}
				for i := offset; i < end; i++ {
					person := people[i]
					result = append(result, []string{
						fmt.Sprintf("%d", person.ID),
						person.Name,
						person.Email,
						person.Status,
					})
				}
				return result
			},
		},
	}

	// Create fynerisor window using NewApp convenience function
	fyneWindow := gui.NewApp("Table Display Example",
		gui.WithGlobals(risor.WithEnv(dataFuncs)),
		gui.WithStatusCallback(func(status string) {
			fmt.Println("Status:", status)
		}),
	)

	// Load Risor script from file
	script, err := os.ReadFile("table.risor")
	if err != nil {
		log.Fatal(err)
	}

	// Load and execute the script
	fyneWindow.LoadScript(string(script))
	fyneWindow.Execute()

	// Show window and run the application
	fyneWindow.ShowAndRun()
}
