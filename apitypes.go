package main

import (
	"fmt"
)

// JSON Object structs.

type AssetCreateSchema struct {
	Title 	string `json:"title"`
}

type FormatSchema struct {
	AssetID 	string `json:"asset_id"`
	Id			string `json:"id"`
	Name		string `json:"name"`
}

type ComponentSchema struct {
	Id		string `json:"id"`
	Name	string `json:"name"`
	Type	string `json:"type"`
}

type FileSetSchema struct {
	Id				string	  `json:"id"`
	BaseDir			string	  `json:"base_dir"`
	ComponentIDs	[]string  `json:"component_ids"`
	FormatID		string	  `json:"format_id"`
	Name			string	  `json:"name"`
}

type FileCreateSchema struct {
	AssetID			string 	`json:"asset_id"`
	FileSetID		string 	`json:"file_set_id"`
	DirectoryPath 	string 	`json:"directory_path"`
	Name			string  `json:"name"`
	OriginalName	string 	`json:"original_name"`
	Type			string 	`json:"type"`
	Size			int64	`json:"size"`
}

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

type AssetSchema struct {
	Id 		string 	 `json:"id"`
}

type SearchResponse struct {
	Objects []IconikObject `json:"objects"`
}

type IconikObject struct {
	Id      string        `json:"id"`
	Files   []IconikFile  `json:"files"`
	Proxies []IconikProxy `json:"proxies"`
}

type IconikProxy struct {
	Id string `json:"id"`
}

type IconikFile struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// ProxyGetUrlSchema is empty. This is because as of 2022Q1, proxies/{proxy_id}
// calls take no arguments in their body.
type ProxyGetUrlSchema struct {
}

type Object struct {
	URL  	 string `json:"url"`
	Type 	 string `json:"type"`
}

type GetResponse struct {
	Objects []Object `json:"objects"`
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
