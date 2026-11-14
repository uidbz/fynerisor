package v3scene

import (
	"errors"
	"strings"
	"sync"
)

// Store is the demo's capability. In ModeHostFunctions its methods are host
// functions called during a run; in ModeActions the runtime calls Customers
// and Save when it performs a fetch or save action. The mutex makes both
// paths safe if a real host performs I/O on a worker goroutine.
type Store struct {
	mu        sync.Mutex
	customers []Customer
}

// NewStore returns a store with a few seed rows.
func NewStore() *Store {
	return &Store{customers: []Customer{
		{Name: "Ada Lovelace", City: "London", Active: true},
		{Name: "Grace Hopper", City: "New York", Active: true},
		{Name: "Alan Turing", City: "Manchester", Active: false},
	}}
}

// Customers returns a copy of the current rows.
func (s *Store) Customers() []Customer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Customer(nil), s.customers...)
}

// Add validates and appends one row. It is the host-function form of a write.
func (s *Store) Add(name, city string) (Customer, error) {
	name, city = strings.TrimSpace(name), strings.TrimSpace(city)
	if name == "" || city == "" {
		return Customer{}, errors.New("name and city are required")
	}
	c := Customer{Name: name, City: city, Active: true}
	s.mu.Lock()
	s.customers = append(s.customers, c)
	s.mu.Unlock()
	return c, nil
}

// Save appends a row supplied by a save action.
func (s *Store) Save(c Customer) {
	s.mu.Lock()
	s.customers = append(s.customers, c)
	s.mu.Unlock()
}
