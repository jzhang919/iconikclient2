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
	Id            string        `json:"id"`
	Title         string        `json:"title"`
	Files         []IconikFile  `json:"files"`
	Proxies       []IconikProxy `json:"proxies"`
	ObjectType    string        `json:"object_type"`
	InCollections []string      `json:"in_collections"` // or parents?
	Status        string        `json:"status"`
}

type CollectionResult struct {
	Path         string `json:"path"`
	CollectionID string `json:"collection_id"`
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
	URL  string `json:"url"`
	Type string `json:"type"`
}

type GetResponse struct {
	Objects []Object `json:"objects"`
}

type AuthToken struct {
	AuthorizationToken string `json:"authorizationToken"`
}

type FileReqResponse struct {
	Id                string    `json:"id"`
	UploadURL         string    `json:"upload_url"`
	UploadCredentials AuthToken `json:"upload_credentials"`
	UploadFilename    string    `json:"upload_filename"`
}

type PostAssetResponse struct {
	Id            string `json:"id"`
	CreatedByUser string `json:"created_by_user"`
}

type NewAssetUpload struct {
	AssetID         string `json:"asset_id"`
	UploadURL       string `json:"upload_url"`
	UploadAuthToken string `json:"upload_auth_token"`
	UploadFilename  string `json:"upload_filename"`
	MimeType        string `json:"mime_type"`
	JobID           string `json:"job_id"`
	FileReqID       string `json:"file_req_id"`
	FileSize        int64  `json:"file_size"`
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
