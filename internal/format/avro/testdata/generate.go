// This file generates test Avro files for testing.
// Run with: go run generate.go

//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/hamba/avro/v2/ocf"
)

func main() {
	// Schema for simple user records
	schemaJSON := `{
		"type": "record",
		"name": "User",
		"fields": [
			{"name": "name", "type": "string"},
			{"name": "age", "type": "int"},
			{"name": "active", "type": "boolean"}
		]
	}`

	// Create users.avro with multiple records
	f, err := os.Create("users.avro")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	enc, err := ocf.NewEncoder(schemaJSON, f)
	if err != nil {
		log.Fatalf("Failed to create encoder: %v", err)
	}
	defer enc.Close()

	users := []map[string]any{
		{"name": "Alice", "age": 30, "active": true},
		{"name": "Bob", "age": 25, "active": false},
		{"name": "Charlie", "age": 35, "active": true},
	}

	for _, user := range users {
		if err := enc.Encode(user); err != nil {
			log.Fatalf("Failed to encode record: %v", err)
		}
	}

	if err := enc.Flush(); err != nil {
		log.Fatalf("Failed to flush encoder: %v", err)
	}

	log.Println("Generated users.avro")

	// Create single.avro with one record
	f2, err := os.Create("single.avro")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f2.Close()

	enc2, err := ocf.NewEncoder(schemaJSON, f2)
	if err != nil {
		log.Fatalf("Failed to create encoder: %v", err)
	}
	defer enc2.Close()

	if err := enc2.Encode(map[string]any{"name": "Solo", "age": 42, "active": true}); err != nil {
		log.Fatalf("Failed to encode record: %v", err)
	}

	if err := enc2.Flush(); err != nil {
		log.Fatalf("Failed to flush encoder: %v", err)
	}

	log.Println("Generated single.avro")
}
