package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackc/pgtype"
	_ "github.com/lib/pq" // postgres driver pq
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "please provide connection string to Postgres database\n")
		return
	}
	connectionString := os.Args[1]
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sql open: %v\n", err)
		return
	}
	defer db.Close()

	query := `SELECT '{1,2}'::text[]`
	fmt.Printf("Marshalling result of `%s` into JSON.\n", query)

	// pgtype
	var ta *pgtype.TextArray
	if err := db.QueryRow(query).Scan(&ta); err != nil {
		fmt.Fprintf(os.Stderr, "query row: %v\n", err)
		return
	}
	arrayJSON, err := json.Marshal(ta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal: %v\n", err)
		return
	}
	fmt.Printf("pgtype: %s\n", string(arrayJSON))

	// pgtype
	var taa *TextArray
	if err := db.QueryRow(query).Scan(&taa); err != nil {
		fmt.Fprintf(os.Stderr, "query row: %v\n", err)
		return
	}
	arrayJSON, err = json.Marshal(taa)
	if err != nil {
		fmt.Fprintf(os.Stderr, "json marshal: %v\n", err)
		return
	}
	fmt.Printf("custom: %s\n", string(arrayJSON))
}

// TextArray wraps pgtype.TextArray, which has support for scanning Postgres arrays
// of arbitrary dimension containing `text` elements.
// Scanning other values (e.g. intervals) into this type also works.
type TextArray struct {
	pgtype.TextArray
}

func (a *TextArray) MarshalJSON() ([]byte, error) {
	elements := make([]interface{}, len(a.Elements))
	for i, e := range a.Elements {
		if notNull(e.Status) {
			elements[i] = e.String
		}
	}
	elementsWithDimensions, err := reconstruct(elements, extractSubDimensions(a.Dimensions))
	if err != nil {
		return nil, fmt.Errorf("reconstructing array failed: %w", err)
	}
	return json.Marshal(elementsWithDimensions)
}

// reconstruct an array of given dimensions from a flat list of its elements.
// The pgtype package we use to scan into does not keep the arrays in their proper structure,
// only as a flat list of elements. But we do have its dimensions. Furthermore,
// the maximum allowed number of array dimensions is 6, and
// multidimensional arrays in Postgres must have matching sub-dimensions for each dimension.
// This allows us to always reconstruct the array.
// Note: Dimensions > 2 left out for brevity
func reconstruct(elements []interface{}, dimensions []int) (interface{}, error) {
	if len(dimensions) == 0 {
		return []interface{}{}, nil
	}
	switch len(dimensions) {
	case 0, 1:
		return elements, nil
	case 2:
		return reconstruct2D(elements, dimensions), nil
	// 3 to 6 left out for brevity
	default:
		return nil, fmt.Errorf("invalid dimension: %d", len(dimensions))
	}
}

func reconstruct2D(elements []interface{}, dimensions []int) [][]interface{} {
	x := make([][]interface{}, dimensions[0])
	for i := 0; i < dimensions[0]; i++ {
		x[i] = make([]interface{}, dimensions[1])
	}
	p := dimensions[1]
	var i0, i1 int
	for i, e := range elements {
		i0 = i / p
		i1 = i % p
		x[i0][i1] = e
	}
	return x
}

func extractSubDimensions(dims []pgtype.ArrayDimension) []int {
	ds := make([]int, len(dims))
	for i, d := range dims {
		ds[i] = int(d.Length)
	}
	return ds
}

// notNull indicates that a value of a pgtype is present
func notNull(status pgtype.Status) bool {
	return status == pgtype.Present
}
