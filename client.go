// Package iconik API for Golang
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	IconikHost               = "https://app.iconik.io/API/"
	searchEndpoint           = "search/v1/search/"
	proxyEndpointTemplate    = "files/v1/assets/%s/proxies"
	fileEndpointTemplate     = "files/v1/assets/%s/files?generate_signed_url=true"
	keyframeEndpointTemplate = "files/v1/assets/%s/keyframes?generate_signed_url=true"
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
		Debug:		 debug,
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

func (c *IClient) parseUrlResponse(resp *http.Response) (string, error) {
	response := GetResponse{}
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

func makeSearchBody(title string, tag string) SearchCriteriaSchema {
	filter := SearchFilter{
		Operator: "OR",
		Terms: []FilterTerm{{
			Name:  "metadata._gcvi_tags",
			Value: tag,
		},
		{
			Name: "title",
			Value: title,
		},
	},
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
	return c.SearchWithTitleAndTag("", tag)
}

// New Function Search With Title:
func(c* IClient) SearchWithTitleAndTag(title string, tag string) (*SearchResponse, error) {
	request := makeSearchBody(title, tag)
	body, err := json.Marshal(request)
	if err != nil {
		return &SearchResponse{}, err
	}
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("Content-Type", "application/json")
	if c.Debug {
		log.Println("----")
		log.Printf("SearchWithTitleAndTag: %s %s %v", searchEndpoint, body, header)
	}
	resp, err := c.post(searchEndpoint, bytes.NewReader(body), header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v) returned an error: %v\n", searchEndpoint, body, err)
		}
		return &SearchResponse{}, err
	}

	return c.parseSearchResponse(resp);
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

	return c.parseUrlResponse(resp)
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

	for _, v := range response.Objects {
		if v.Type == "KEYFRAME" {
			return v.URL, nil
		}
	}

	return "", &IError{Errors: []string{"didn't find KEYFRAME"}}
}
