// Package iconik API for Golang
package iconik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	iconikHost            = "https://app.iconik.io/API/"
	searchEndpoint        = "search/v1/search/"
	proxyEndpointTemplate = "files/v1/assets/%s/proxies/%s/download_url/"
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
func NewIClient(creds Credentials, host string) (*IClient, error) {
	if host == "" {
		host = iconikHost
	}
	c := &IClient{
		Credentials: creds,
		host:        iconikHost,
	}

	return c, nil
}

// Create an authorized request using the client's credentials
func (c *IClient) newRequest(method, apiPath string, body io.Reader, headerSettings http.Header) (*http.Request, error) {
	path := c.host + apiPath
	header := make(http.Header)
	header.Add("App-Id", c.AppID)
	header.Add("Auth-Token", c.Token)
	for k, vs := range headerSettings {
		for _, value := range vs {
			if c.Debug {
				log.Printf("Adding (%s, %s) to header\n", k, value)
			}
			header.Add(k, value)
		}
	}

	req, err := http.NewRequest(method, path, body)
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

func (c *IClient) post(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {
	req, err := c.newRequest(http.MethodPost, apiPath, body, header)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	return resp, err
}

// Attempts to parse a response body into the provided result struct

func (c *IClient) parseSearchResponse(resp *http.Response) (*SearchResponse, error) {
	response := SearchResponse{}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
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

func (c *IClient) parseProxyUrlResponse(resp *http.Response) (string, error) {
	response := ProxyGetUrlResponse{}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
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

func makeSearchBody(tag string) SearchCriteriaSchema {
	filter := SearchFilter{
		Operator: "AND",
		Terms: []FilterTerm{{
			Name:  "metadata._gcvi_tags",
			Value: tag,
		}},
	}
	schema := SearchCriteriaSchema{
		DocTypes: []string{"assets"},
		Filter:   filter,
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
func (c *IClient) SearchWithTag(tag string) (*SearchResponse, error) {
	request := makeSearchBody(tag)
	body, err := json.Marshal(request)
	if err != nil {
		return &SearchResponse{}, err
	}
	header := make(http.Header)
  header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	if c.Debug {
		log.Println("----")
		log.Printf("SearchWithTag: %s %s %v", searchEndpoint, body, header)
	}

	resp, err := c.post(searchEndpoint, bytes.NewReader(body), header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", searchEndpoint, body, err)
		}
		return &SearchResponse{}, err
	}

	return c.parseSearchResponse(resp)
}

func (c *IClient) GenerateSignedProxyUrl(assetID, proxyID string) (string, error) {
	header := make(http.Header)
	header.Add("asset_id", assetID)
	header.Add("proxy_id", proxyID)
	proxyEndpoint := fmt.Sprintf(proxyEndpointTemplate, assetID, proxyID)
	if c.Debug {
		log.Println("----")
		log.Printf("GenerateSignedProxyUrl: %s %s %v", proxyEndpoint, nil, header)
	}
	resp, err := c.get(proxyEndpoint, nil, header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.get(%s, %v) returned an error: %v\n", proxyEndpoint, nil, err)
		}
		return "", err
	}
	return c.parseProxyUrlResponse(resp)
}
