// Package iconik API for Golang
package iconik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	MULTIPART_FILESIZE_THRESHOLD      = 100 * 1024 * 1024 // 100MB
	IconikHost                        = "https://app.iconik.io/API/"
	searchEndpoint                    = "search/v1/search/"
	collectionEndpointTemplate        = "assets/v1/collections/%s"
	proxyEndpointTemplate             = "files/v1/assets/%s/proxies"
	fileEndpointTemplate              = "files/v1/assets/%s/files"
	fileEndpointTemplate2             = "files/v1/assets/%s/files/%s/download_url"
	keyframeEndpointTemplate          = "files/v1/assets/%s/keyframes?generate_signed_url=true"
	postAssetEndpointTemplate         = "assets/v1/assets?assign_to_collection=true"
	storagesMatchingEndpoint          = "files/v1/storages/matching/FILES"
	formatIDEndpointTemplate          = "files/v1/assets/%s/formats"
	filesetsEndpointTemplate          = "files/v1/assets/%s/file_sets"
	uploadUrlEndpointTemplate         = "files/v1/assets/%s/files/"
	multipartStartEndpointTemplate    = "files/v1/assets/%s/files/%s/multipart/b2/start/"
	multipartFinishEndpointTemplate   = "files/v1/assets/%s/files/%s/multipart/b2/finish/"
	jobStartEndpointTemplate          = "jobs/v1/jobs"
	uploadUrlFinishedEndpointTemplate = "files/v1/assets/%s/files/%s/"
	keyframeGenerateEndpointTemplate  = "files/v1/assets/%s/files/%s/keyframes/"
	patchJobCompleteEndpointTemplate  = "jobs/v1/jobs/%s"
)

// Credentials are the identification required by the Iconik API
//
// The AppID is the application key id that you
// get when generating an application key.
//
// The Token is the string generated alongside the
// application key.
type Credentials struct {
	AppID string
	Token string
}

// IClient implements a Iconik client. Do not modify state concurrently.
type IClient struct {
	Credentials

	// If true, don't retry requests if authorization has expired
	NoRetry bool

	// If true, display debugging information about API calls
	Debug bool

	// State
	host       string
	httpClient http.Client
}

// NewIClient creates a new Client for accessing the Iconik API.
func NewIClient(creds Credentials, host string, debug bool) (*IClient, error) {
	if host == "" {
		host = IconikHost
	} else if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	c := &IClient{
		Credentials: creds,
		host:        host,
		Debug:       debug,
	}

	return c, nil
}

// Create an authorized request using the client's credentials
func (c *IClient) newRequest(method, apiPath string, body io.Reader, headerSettings http.Header) (*http.Request, error) {
	path := c.host + apiPath
	header := make(http.Header)
	header.Add("App-Id", c.AppID)
	header.Add("Auth-Token", c.Token)
	header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	for k, vs := range headerSettings {
		for _, value := range vs {
			if c.Debug {
				log.Printf("Adding (%s, %s) to header\n", k, value)
			}
			header.Add(k, value)
		}
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header = header

	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("newRequest: %s %s\n", method, req.URL)
	}

	return req, nil
}

// Dispatch an authorized API GET request
func (c *IClient) get(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := c.newRequest(http.MethodGet, apiPath, body, header)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	return resp, err
}

// Dispatch an authorized API POST request
func (c *IClient) post(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := c.newRequest(http.MethodPost, apiPath, body, header)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	return resp, err
}

// Dispatch an authorized API PATCH request
func (c *IClient) patch(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := c.newRequest(http.MethodPatch, apiPath, body, header)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	return resp, err
}

// Attempts to parse a response body
func (c *IClient) parseSearchResponse(resp *http.Response) (*SearchResponse, error) {
	response := SearchResponse{}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	// Check response code
	switch resp.StatusCode {
	case 200: // Response is OK
	default:
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return nil, &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return nil, iErr
	}
	err = json.Unmarshal(body, &response)
	return &response, err
}

func (c *IClient) parseUrlResponse(resp *http.Response) (string, error) {
	response := GetResponse{}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	// Check response code
	switch resp.StatusCode {
	case 200: // Response is OK
	default:
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return "", iErr
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	retVal := ""
	for _, v := range response.Objects {
		if v.URL != "" {
			if retVal != "" {
				return "", fmt.Errorf("more than one URL in response")
			}
			retVal = v.URL
		}
	}

	return retVal, nil
}

func (c *IClient) parseObjectResponse(resp *http.Response) (string, error) {
	response := Object{}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	// Check response code
	switch resp.StatusCode {
	case 200: // Response is OK
	default:
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return "", iErr
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.URL, nil
}

func makeSearchBody(title string, tag string, isCollection bool) SearchCriteriaSchema {
	filter := SearchFilter{
		Operator: "OR",
		Terms: []FilterTerm{{
			Name:  "metadata._gcvi_tags",
			Value: tag,
		},
			{
				Name:  "title",
				Value: title,
			},
		},
	}
	schema := SearchCriteriaSchema{
		DocTypes: []string{"assets"},
		Filter:   filter,
	}
	if isCollection {
		schema.DocTypes = []string{"collections"}
	}
	return schema
}

func makeProxyUrlBody() ProxyGetUrlSchema {
	return ProxyGetUrlSchema{}
}

// SearchWithTag performs an Iconik API Search for assets with the matching tag.
// Args:
// apiPath: The API Resource
// tag: The metadata tag on Iconik, eg. "TeachingVideos," that you want to find matching assets for.
// response: The response object to be filled out.
func (c *IClient) SearchWithTag(tag string, isCollection bool) (*SearchResponse, error) {
	return c.SearchWithTitleAndTag("", tag, isCollection)
}

func (c *IClient) GetKeyframeUrl(assetID string) (string, error) {
	header := make(http.Header)
	header.Add("asset_id", assetID)
	keyframeEndpoint := fmt.Sprintf(keyframeEndpointTemplate, assetID)
	if c.Debug {
		log.Println("----")
		log.Printf("GetKeyframeUrl: %s %v", keyframeEndpoint, header)
	}
	resp, err := c.get(keyframeEndpoint, nil, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s) returned an error: %v\n", keyframeEndpoint, err)
		}
		return "", err
	}

	response := GetResponse{}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	// Check response code
	switch resp.StatusCode {
	case 200: // Response is OK
	default:
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return "", iErr
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	for _, v := range response.Objects {
		if v.Type == "KEYFRAME" {
			return v.URL, nil
		}
	}

	return "", &IError{Errors: []string{"didn't find KEYFRAME"}}
}

// New Function Search With Title:
func (c *IClient) SearchWithTitleAndTag(title string, tag string, isCollection bool) (*SearchResponse, error) {
	request := makeSearchBody(title, tag, isCollection)
	body, err := json.Marshal(request)
	if err != nil {
		return &SearchResponse{}, err
	}
	if c.Debug {
		log.Println("----")
		log.Printf("SearchWithTitleAndTag: %s %s", searchEndpoint, body)
	}
	resp, err := c.post(searchEndpoint, bytes.NewReader(body), http.Header{})
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", searchEndpoint, body, err)
		}
		return &SearchResponse{}, err
	}

	return c.parseSearchResponse(resp)
}

func (c *IClient) GenerateSignedProxyUrl(assetID string) (string, error) {
	header := make(http.Header)
	header.Add("asset_id", assetID)
	proxyEndpoint := fmt.Sprintf(proxyEndpointTemplate, assetID)
	if c.Debug {
		log.Println("----")
		log.Printf("GenerateSignedProxyUrl: %s %v", proxyEndpoint, header)
	}
	resp, err := c.get(proxyEndpoint, nil, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s) returned an error: %v\n", proxyEndpoint, err)
		}
		return "", err
	}

	return c.parseUrlResponse(resp)
}

// the documentation says that you can just add the query parameter "generate_signed_url=true" to the fileEndpointTemplate
// but when you do that, you get a download URL which, when going to fetch it, doesn't have correct content-disposition nor filename
// so you have to do this roundabout way instead: first get the fileID, then call fileEndpointTemplate2 which gives you the download URL
func (c *IClient) GenerateSignedFileUrl(assetID string) (string, error) {
	header := make(http.Header)
	header.Add("asset_id", assetID)
	fileEndpoint := fmt.Sprintf(fileEndpointTemplate, assetID)
	if c.Debug {
		log.Println("----")
		log.Printf("GenerateSignedFileUrl: %s %v", fileEndpoint, header)
	}
	resp, err := c.get(fileEndpoint, nil, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s) returned an error: %v\n", fileEndpoint, err)
		}
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type Response struct {
		Objects []struct {
			ID string `json:"id"` // this is the fileID
		}
	}
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		log.Fatalf("Error decoding JSON: %v", err)
	}

	// now that we have the fileID, go and get the signed URL
	fileEndpoint2 := fmt.Sprintf(fileEndpointTemplate2, assetID, r.Objects[0].ID)
	if c.Debug {
		log.Println("----")
		log.Printf("GenerateSignedFileUrl: %s %v", fileEndpoint, header)
	}
	resp, err = c.get(fileEndpoint2, nil, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s) returned an error: %v\n", fileEndpoint, err)
		}
		return "", err
	}
	return c.parseObjectResponse(resp)
}

func (c *IClient) postAndGetID(endpoint string, reqBody io.Reader, header http.Header) (string, error) {
	resp, err := c.post(endpoint, reqBody, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", endpoint, reqBody, err)
		}
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if c.Debug {
		log.Printf("Response: %s", body)
	}
	if resp.StatusCode != 201 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return "", iErr
	}

	type IDResponse struct {
		Id string `json:"id"`
	}
	idResponse := IDResponse{}
	err = json.Unmarshal(body, &idResponse)
	if err != nil {
		return "", err
	}
	return idResponse.Id, nil
}

func (c *IClient) patchAndGetID(endpoint string, reqBody io.Reader, header http.Header) (string, error) {
	req, err := c.newRequest(http.MethodPatch, endpoint, reqBody, header)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", endpoint, reqBody, err)
		}
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if c.Debug {
		log.Printf("Response: %s", body)
	}
	if resp.StatusCode != 201 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return "", iErr
	}

	type IDResponse struct {
		Id string `json:"id"`
	}
	idResponse := IDResponse{}
	err = json.Unmarshal(body, &idResponse)
	if err != nil {
		return "", err
	}
	return idResponse.Id, nil
}

func (c *IClient) getAndGetID(endpoint string, reqBody io.Reader, header http.Header) (string, error) {
	resp, err := c.get(endpoint, reqBody, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s) returned an error: %v\n", endpoint, err)
		}
		return "", err
	}
	defer resp.Body.Close()
	type IDResponse struct {
		Id string `json:"id"`
	}
	idResponse := IDResponse{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return "", &IError{
				Errors: []string{"UNKNOWN: error message nor parsable"},
			}
		}
		return "", iErr
	}
	err = json.Unmarshal(body, &idResponse)
	if err != nil {
		return "", err
	}
	return idResponse.Id, nil
}

// GetCollectionID will return the ID of the collection with the given name.
func (c IClient) GetCollectionIDs(collectionName string) ([]*CollectionResult, error) {
	collectionResp, err := c.SearchWithTitleAndTag(collectionName, "", true)
	if err != nil {
		return nil, err
	}
	var getColPath func(string) (string, error)
	getColPath = func(collectionID string) (string, error) {
		collectionEndpoint := fmt.Sprintf(collectionEndpointTemplate, collectionID)
		resp, err := c.get(collectionEndpoint, nil, http.Header{})
		if err != nil {
			return "", err
		}
		response := IconikObject{}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		if c.Debug {
			log.Printf("Response: %s", body)
		}

		// Check response code
		switch resp.StatusCode {
		case 200: // Response is OK
		case 404:
			return "", nil
		default:
			iErr := &IError{}
			if err := json.Unmarshal(body, iErr); err != nil {
				if c.Debug {
					log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
				}
				return "", &IError{
					Errors: []string{"UNKNOWN; error message not parsable"},
				}
			}
			return "", iErr
		}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return "", err
		}
		if response.Status == "DELETED" {
			return "", nil
		}
		if len(response.InCollections) == 0 {
			return response.Title, nil
		}
		next, err := getColPath(response.InCollections[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s/%s", next, response.Title), nil
	}

	retVal := []*CollectionResult{}
	for _, col := range collectionResp.Objects {
		for _, ancestor := range col.InCollections {
			cp, err := getColPath(ancestor)
			if err != nil {
				return nil, err
			}
			if len(cp) == 0 {
				continue
			}
			retVal = append(retVal, &CollectionResult{
				Path:         fmt.Sprintf("%s/%s", cp, col.Title),
				CollectionID: col.Id,
			})
		}
	}

	return retVal, nil
}

// PostAssetID will create an asset with the given title in the given collection.
func (c IClient) PostAssetID(collectionID, title string) (*PostAssetResponse, error) {
	endpoint := fmt.Sprintf(postAssetEndpointTemplate)
	reqBody := map[string]string{
		"collection_id": collectionID,
		"title":         title,
	}
	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	resp, err := c.post(endpoint, bytes.NewReader(reqBodyJSON), header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", endpoint, reqBodyJSON, err)
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		log.Printf("Response: %s", body)
	}

	if resp.StatusCode != 201 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return nil, &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return nil, iErr
	}

	postAssetResponse := PostAssetResponse{}
	err = json.Unmarshal(body, &postAssetResponse)
	if err != nil {
		return nil, err
	}

	return &postAssetResponse, nil
}

// MakeStorageID will create a storage ID for the asset.
func (c *IClient) MakeStorageID() (string, error) {
	// now make the storageID
	storageID, err := c.getAndGetID(storagesMatchingEndpoint, nil, http.Header{})
	if err != nil {
		return "", err
	}
	return storageID, nil
}

// MakeFormatID will create a format ID for the asset.
func (c *IClient) MakeFormatID(userID, assetID, mimeType string) (string, error) {
	// now make the formatID
	endpoint := fmt.Sprintf(formatIDEndpointTemplate, assetID)
	type IMD struct {
		InternetMediaType string `json:"internet_media_type"`
	}
	type FormatIDReq struct {
		UserId   string `json:"user_id"`
		Name     string `json:"name"`
		Metadata []IMD  `json:"metadata"`
	}
	formatIDReqBody := FormatIDReq{
		UserId:   userID,
		Name:     "ORIGINAL",
		Metadata: []IMD{IMD{mimeType}},
	}
	reqBodyJSON, err := json.Marshal(formatIDReqBody)
	if err != nil {
		return "", err
	}
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	formatID, err := c.postAndGetID(endpoint, bytes.NewReader(reqBodyJSON), header)
	if err != nil {
		return "", err
	}
	return formatID, nil
}

// MakeFileSetID will create a fileset ID for the asset.
func (c *IClient) MakeFileSetID(assetID, formatID, storageID, title, baseDir string) (string, error) {
	endpoint := fmt.Sprintf(filesetsEndpointTemplate, assetID)
	type FileSetIDReq struct {
		FormatID     string   `json:"format_id"`
		StorageID    string   `json:"storage_id"`
		BaseDir      string   `json:"base_dir"`
		Name         string   `json:"name"`
		ComponentIDs []string `json:"component_ids"`
	}
	fileSetReqBody := FileSetIDReq{
		FormatID:     formatID,
		StorageID:    storageID,
		BaseDir:      baseDir,
		Name:         title,
		ComponentIDs: []string{},
	}
	reqBodyJSON, err := json.Marshal(fileSetReqBody)
	if err != nil {
		return "", err
	}
	fileSetID, err := c.postAndGetID(endpoint, bytes.NewReader(reqBodyJSON), http.Header{})
	if err != nil {
		return "", err
	}
	return fileSetID, nil
}

// GetUploadUrl will get the upload URL for the asset. (only works for BackBlaze right now)
func (c *IClient) GetUploadUrl(assetID, title, directoryPath, formatID, fileSetID, storageID, fileDateCreated string, fileSize int64) (*FileReqResponse, error) {
	endpoint := fmt.Sprintf(uploadUrlEndpointTemplate, assetID)
	type FileReq struct {
		OriginalName     string `json:"original_name"`
		DirectoryPath    string `json:"directory_path"`
		Size             int64  `json:"size"`
		Type             string `json:"type"`
		Metadata         string `json:"metadata"`
		FormatID         string `json:"format_id"`
		FileSetID        string `json:"file_set_id"`
		StorageID        string `json:"storage_id"`
		FileDateCreated  string `json:"file_date_created"`
		FileDateModified string `json:"file_date_modified"`
	}
	fileReqBody := FileReq{
		OriginalName:     title,
		DirectoryPath:    directoryPath,
		Size:             fileSize,
		Type:             "FILE",
		Metadata:         "{}",
		FormatID:         formatID,
		FileSetID:        fileSetID,
		StorageID:        storageID,
		FileDateCreated:  fileDateCreated,
		FileDateModified: fileDateCreated,
	}
	reqBodyJSON, err := json.Marshal(fileReqBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.post(endpoint, bytes.NewReader(reqBodyJSON), http.Header{})
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s) returned an error: %v\n", endpoint, err)
		}
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		log.Printf("Response: %s", body)
	}
	if resp.StatusCode != 201 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return nil, &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return nil, iErr
	}

	frResponse := FileReqResponse{}
	err = json.Unmarshal(body, &frResponse)
	if err != nil {
		return nil, err
	}
	if c.Debug {
		log.Printf("Upload URL response: %s", string(body))
	}
	return &frResponse, nil
}

func (c *IClient) GetMultipartStartUrl(NAU *NewAssetUpload) error {
	endpoint := fmt.Sprintf(multipartStartEndpointTemplate, NAU.AssetID, NAU.FileReqID)
	type MultipartStartReq struct {
		AssetID string `json:"asset_id"`
		FileID  string `json:"file_id"`
	}
	multiStartReqBody := MultipartStartReq{
		AssetID: NAU.AssetID,
		FileID:  NAU.FileReqID,
	}
	reqBodyJSON, err := json.Marshal(multiStartReqBody)
	if err != nil {
		return err
	}
	resp, err := c.post(endpoint, bytes.NewReader(reqBodyJSON), http.Header{})
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s) returned an error: %v\n", endpoint, err)
		}
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return iErr
	}

	type MultipartStartResp struct {
		AuthorizationToken string `json:"authorization_token"`
		UploadFileID       string `json:"upload_file_id"`
		UploadURL          string `json:"upload_url"`
	}

	frResponse := MultipartStartResp{}
	err = json.Unmarshal(body, &frResponse)
	if err != nil {
		return err
	}
	if c.Debug {
		log.Printf("Upload URL response: %s", string(body))
	}
	NAU.UploadAuthToken = frResponse.AuthorizationToken
	NAU.UploadURL = frResponse.UploadURL
	NAU.MultipartFileID = frResponse.UploadFileID
	return nil
}

// PostStartOfJob will post the start of a job.
func (c *IClient) PostStartOfJob(assetID, title string) (string, error) {
	type JobReq struct {
		ObjectType string `json:"object_type"`
		ObjectID   string `json:"object_id"`
		Type       string `json:"type"`
		Status     string `json:"status"`
		Title      string `json:"title"`
	}
	jobReqBody := JobReq{
		ObjectType: "assets",
		ObjectID:   assetID,
		Type:       "TRANSFER",
		Status:     "STARTED",
		Title:      title,
	}
	reqBodyJSON, err := json.Marshal(jobReqBody)
	if err != nil {
		return "", err
	}
	jobID, err := c.postAndGetID(jobStartEndpointTemplate, bytes.NewReader(reqBodyJSON), http.Header{})
	if err != nil {
		return "", err
	}
	if c.Debug {
		log.Printf("jobID: %s", jobID)
	}
	return jobID, nil
}

// MakeNewAsset will create a new asset with the given title in the given collection.
// This ties together many steps needed to actually upload a file. Ultimately, it returns
// a NewAssetUpload object that contains all the information needed to upload a file. Once
// done, you can call FinishUpload to finish the upload.
func (c *IClient) MakeNewAsset(collectionID, fileName, title, storagePath, mimeType string, fileSize int64, fileDateCreated time.Time) (*NewAssetUpload, error) {
	NAU := &NewAssetUpload{
		MimeType: mimeType,
		FileSize: fileSize,
	}

	// create the Asset
	postAssetResponse, err := c.PostAssetID(collectionID, title)
	if err != nil {
		return nil, err
	}
	NAU.AssetID = postAssetResponse.Id

	// now make the storageID
	storageID, err := c.MakeStorageID()
	if err != nil {
		return nil, err
	}

	// now make the formatID
	formatID, err := c.MakeFormatID(postAssetResponse.CreatedByUser, postAssetResponse.Id, mimeType)
	if err != nil {
		return nil, err
	}
	// now the filesetID
	fileSetId, err := c.MakeFileSetID(postAssetResponse.Id, formatID, storageID, title, storagePath)
	if err != nil {
		return nil, err
	}

	// get upload URL
	frResponse, err := c.GetUploadUrl(postAssetResponse.Id, title, storagePath, formatID, fileSetId, storageID, fileDateCreated.Format(time.RFC3339), fileSize)
	if err != nil {
		return nil, err
	}
	NAU.UploadURL = frResponse.UploadURL
	NAU.UploadAuthToken = frResponse.UploadCredentials.AuthorizationToken
	NAU.UploadFilename = frResponse.UploadFilename
	NAU.FileReqID = frResponse.Id

	if fileSize > MULTIPART_FILESIZE_THRESHOLD {
		if err := c.GetMultipartStartUrl(NAU); err != nil {
			return nil, err
		}
	}

	// Note start of job
	jobID, err := c.PostStartOfJob(postAssetResponse.Id, title)
	if err != nil {
		return nil, err
	}
	NAU.JobID = jobID

	return NAU, nil
}

// CloseFileRequest will close the file request.
func (c *IClient) CloseFileRequest(assetID, fileReqID string) error {
	endpoint := fmt.Sprintf(uploadUrlFinishedEndpointTemplate, assetID, fileReqID)
	type FinishedReq struct {
		Status            string `json:"status"`
		ProgressProcessed int    `json:"progress_processed"`
	}
	finishedReqBody := FinishedReq{
		Status:            "CLOSED",
		ProgressProcessed: 100,
	}
	reqBodyJSON, err := json.Marshal(finishedReqBody)
	if err != nil {
		return err
	}
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	resp, err := c.patch(endpoint, bytes.NewReader(reqBodyJSON), header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return iErr
	}
	return nil
}

// GenerateKeyframes will generate keyframes for the asset.
func (c *IClient) GenerateKeyframes(assetID, fileReqID string) error {
	endpoint := fmt.Sprintf(keyframeGenerateEndpointTemplate, assetID, fileReqID)
	resp, err := c.post(endpoint, nil, http.Header{})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return iErr
	}
	return nil
}

// FinishJob will finish the job.
func (c *IClient) FinishJob(jobID string) error {
	endpoint := fmt.Sprintf(patchJobCompleteEndpointTemplate, jobID)
	type FinishedJobReq struct {
		Status            string `json:"status"`
		ProgressProcessed int    `json:"progress_processed"`
	}
	finishedJobReqBody := FinishedJobReq{
		Status:            "FINISHED",
		ProgressProcessed: 100,
	}
	reqBodyJSON, err := json.Marshal(finishedJobReqBody)
	if err != nil {
		return err
	}
	resp, err := c.patch(endpoint, bytes.NewReader(reqBodyJSON), http.Header{})
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if c.Debug {
		log.Printf("Response: %s", body)
	}
	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return iErr
	}
	return err
}

func (c *IClient) FinishMultipartUpload(newAssetUpload *NewAssetUpload) error {
	if len(newAssetUpload.Sha1List) == 0 {
		return fmt.Errorf("no sha1 list provided")
	}

	endpoint := fmt.Sprintf(multipartFinishEndpointTemplate, newAssetUpload.AssetID, newAssetUpload.FileReqID)
	type FinishMultipartReq struct {
		Sha1List     []string `json:"sha1_list"`
		UploadFileID string   `json:"upload_file_id"`
	}
	finishMultipartReqBody := FinishMultipartReq{
		Sha1List:     newAssetUpload.Sha1List,
		UploadFileID: newAssetUpload.MultipartFileID,
	}
	reqBodyJSON, err := json.Marshal(finishMultipartReqBody)
	if err != nil {
		return err
	}

	resp, err := c.post(endpoint, bytes.NewReader(reqBodyJSON), http.Header{})
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	if resp.StatusCode != 200 {
		iErr := &IError{}
		if err := json.Unmarshal(body, iErr); err != nil {
			if c.Debug {
				log.Printf("Unmarshal(%v) got %v, wanted to parse", body, err)
			}
			return &IError{
				Errors: []string{"UNKNOWN; error message not parsable"},
			}
		}
		return iErr
	}
	return nil
}

// FinishUpload will finish the upload. (call it after uploading the file), uses previously
// defined steps.
func (c *IClient) FinishUpload(newAssetUpload *NewAssetUpload) error {
	if newAssetUpload.MultipartFileID != "" {
		if err := c.FinishMultipartUpload(newAssetUpload); err != nil {
			return err
		}
	}

	// patch files
	if err := c.CloseFileRequest(newAssetUpload.AssetID, newAssetUpload.FileReqID); err != nil {
		return err
	}

	// generate keyframes
	if err := c.GenerateKeyframes(newAssetUpload.AssetID, newAssetUpload.FileReqID); err != nil {
		return err
	}

	// patch job
	if err := c.FinishJob(newAssetUpload.JobID); err != nil {
		return err
	}
	return nil
}
