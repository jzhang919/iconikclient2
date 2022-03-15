package iconik

import (
	"fmt"
)

// JSON Object structs.

type SearchCriteriaSchema struct {
	DocTypes []string     `json:"doc_types"`
	Filter   SearchFilter `json:"filter"`
}

type SearchFilter struct {
	Operator string       `json:"operator"`
	Terms    []FilterTerm `json:"terms"`
}

type FilterTerm struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SearchResponse struct {
	Objects []IconikObject `json:"objects"`
}

type IconikObject struct {
	Files []IconikFile `json:"files"`
}

type IconikFile struct {
	Name string `json:"name"`
}

// IError encapsulates an error message returned by the Iconik API.
//
// Failures to connect to the Iconik servers, and networking problems in general can cause errors
type IError struct {
	Errors []string `json:"errors"`
}

func (e IError) Error() string {
	return fmt.Sprintf("%v", e.Errors)
}
