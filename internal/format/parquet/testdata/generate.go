// This file generates test Parquet files for testing.
// Run with: go run generate.go

//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/parquet-go/parquet-go"
)

// User represents a user record
type User struct {
	Name   string `parquet:"name"`
	Age    int32  `parquet:"age"`
	Active bool   `parquet:"active"`
}

func main() {
	// Create users.parquet with multiple records
	createUsersFile()

	// Create single.parquet with one record
	createSingleFile()

	log.Println("All parquet files generated successfully")
}

func createUsersFile() {
	f, err := os.Create("users.parquet")
	if err != nil {
		log.Fatalf("Failed to create users.parquet: %v", err)
	}
	defer f.Close()

	users := []User{
		{Name: "Alice", Age: 30, Active: true},
		{Name: "Bob", Age: 25, Active: false},
		{Name: "Charlie", Age: 35, Active: true},
	}

	err = parquet.Write(f, users)
	if err != nil {
		log.Fatalf("Failed to write users: %v", err)
	}

	log.Println("Generated users.parquet")
}

func createSingleFile() {
	f, err := os.Create("single.parquet")
	if err != nil {
		log.Fatalf("Failed to create single.parquet: %v", err)
	}
	defer f.Close()

	user := []User{
		{Name: "Solo", Age: 42, Active: true},
	}

	err = parquet.Write(f, user)
	if err != nil {
		log.Fatalf("Failed to write user: %v", err)
	}

	log.Println("Generated single.parquet")
}
