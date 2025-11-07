//go:build ignore

package main

import (
	"log"
	"os"

	"github.com/hamba/avro/v2/ocf"
	"github.com/parquet-go/parquet-go"
)

// Employee represents an employee record
type Employee struct {
	ID         int    `avro:"id" parquet:"id"`
	Name       string `avro:"name" parquet:"name"`
	Department string `avro:"department" parquet:"department"`
	Salary     int    `avro:"salary" parquet:"salary"`
	Active     bool   `avro:"active" parquet:"active"`
}

// Product represents a product record
type Product struct {
	SKU       string  `parquet:"sku"`
	Name      string  `parquet:"name"`
	Category  string  `parquet:"category"`
	Price     float64 `parquet:"price"`
	InStock   bool    `parquet:"in_stock"`
}

func main() {
	log.Println("Generating comprehensive test data files...")

	// Generate Avro files
	generateEmployees1Avro()
	generateEmployees2Avro()

	// Generate Parquet files
	generateProducts1Parquet()
	generateProducts2Parquet()

	log.Println("Test data generation complete!")
}

func generateEmployees1Avro() {
	employees := []Employee{
		{ID: 1, Name: "Alice Johnson", Department: "Engineering", Salary: 95000, Active: true},
		{ID: 2, Name: "Bob Smith", Department: "Engineering", Salary: 85000, Active: true},
		{ID: 3, Name: "Carol White", Department: "Sales", Salary: 75000, Active: true},
		{ID: 4, Name: "David Brown", Department: "Marketing", Salary: 70000, Active: false},
		{ID: 5, Name: "Eve Davis", Department: "Engineering", Salary: 90000, Active: true},
	}

	writeAvroFile("employees1.avro", employees)
}

func generateEmployees2Avro() {
	employees := []Employee{
		{ID: 6, Name: "Frank Miller", Department: "Sales", Salary: 72000, Active: true},
		{ID: 7, Name: "Grace Lee", Department: "Engineering", Salary: 98000, Active: true},
		{ID: 8, Name: "Henry Wilson", Department: "Marketing", Salary: 68000, Active: true},
		{ID: 9, Name: "Ivy Martinez", Department: "Sales", Salary: 76000, Active: false},
		{ID: 10, Name: "Jack Anderson", Department: "Engineering", Salary: 92000, Active: true},
	}

	writeAvroFile("employees2.avro", employees)
}

func writeAvroFile(filename string, employees []Employee) {
	schema := `{
		"type": "record",
		"name": "Employee",
		"fields": [
			{"name": "id", "type": "int"},
			{"name": "name", "type": "string"},
			{"name": "department", "type": "string"},
			{"name": "salary", "type": "int"},
			{"name": "active", "type": "boolean"}
		]
	}`

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer file.Close()

	encoder, err := ocf.NewEncoder(schema, file, ocf.WithCodec(ocf.Null))
	if err != nil {
		log.Fatalf("Failed to create encoder for %s: %v", filename, err)
	}
	defer encoder.Close()

	for _, emp := range employees {
		empMap := map[string]any{
			"id":         emp.ID,
			"name":       emp.Name,
			"department": emp.Department,
			"salary":     emp.Salary,
			"active":     emp.Active,
		}
		if err := encoder.Encode(empMap); err != nil {
			log.Fatalf("Failed to encode employee in %s: %v", filename, err)
		}
	}

	if err := encoder.Flush(); err != nil {
		log.Fatalf("Failed to flush encoder for %s: %v", filename, err)
	}

	log.Printf("Created %s with %d records", filename, len(employees))
}

func generateProducts1Parquet() {
	products := []Product{
		{SKU: "LAPTOP-001", Name: "Dell XPS 15", Category: "Electronics", Price: 1499.99, InStock: true},
		{SKU: "LAPTOP-002", Name: "MacBook Pro", Category: "Electronics", Price: 2399.99, InStock: true},
		{SKU: "CHAIR-001", Name: "Office Chair Pro", Category: "Furniture", Price: 299.99, InStock: true},
		{SKU: "DESK-001", Name: "Standing Desk", Category: "Furniture", Price: 599.99, InStock: false},
		{SKU: "MONITOR-001", Name: "4K Display", Category: "Electronics", Price: 449.99, InStock: true},
	}

	writeParquetFile("products1.parquet", products)
}

func generateProducts2Parquet() {
	products := []Product{
		{SKU: "KEYBOARD-001", Name: "Mechanical Keyboard", Category: "Electronics", Price: 149.99, InStock: true},
		{SKU: "MOUSE-001", Name: "Ergonomic Mouse", Category: "Electronics", Price: 79.99, InStock: true},
		{SKU: "LAMP-001", Name: "Desk Lamp", Category: "Furniture", Price: 49.99, InStock: false},
		{SKU: "TABLET-001", Name: "iPad Pro", Category: "Electronics", Price: 1099.99, InStock: true},
		{SKU: "HEADPHONES-001", Name: "Noise Canceling", Category: "Electronics", Price: 349.99, InStock: true},
	}

	writeParquetFile("products2.parquet", products)
}

func writeParquetFile(filename string, products []Product) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create %s: %v", filename, err)
	}
	defer file.Close()

	writer := parquet.NewGenericWriter[Product](file)

	if _, err := writer.Write(products); err != nil {
		log.Fatalf("Failed to write products to %s: %v", filename, err)
	}

	if err := writer.Close(); err != nil {
		log.Fatalf("Failed to close parquet writer for %s: %v", filename, err)
	}

	log.Printf("Created %s with %d records", filename, len(products))
}
