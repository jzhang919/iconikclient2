// Package iconik API for Golang
package iconik

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

const (
	iconikHost = "https://app.iconik.io/API/"
	v1         = "/search/v1/"
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
	mutex      sync.Mutex
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
func (c *IClient) authRequest(method, apiPath string, body io.Reader, header http.Header) (*http.Request, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	path := c.host + apiPath
	req, err := http.NewRequest(method, path, body)
	req.Header = header

	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("authRequest: %s %s\n", method, req.URL)
	}

	return req, nil
}

// Dispatch an authorized API GET request
func (c *IClient) authGet(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {

	req, err := c.authRequest(http.MethodGet, apiPath, body, header)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	return resp, err
}

func (c *IClient) authPost(apiPath string, body io.Reader, header http.Header) (*http.Response, error) {

	req, err := c.authRequest(http.MethodPost, apiPath, body, header)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	return resp, err
}

// Looks for an error message in the response body and parses it into a
// IError object
func (c *IClient) parseError(body []byte) error {
	err := &IError{}
	if json.Unmarshal(body, err) != nil {
		return nil
	}
	return err
}

// Attempts to parse a response body into the provided result struct
func (c *IClient) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if c.Debug {
		log.Printf("Response: %s", body)
	}

	// Check response code
	switch resp.StatusCode {
	case 200: // Response is OK
	case 401:
		if err := c.parseError(body); err != nil {
			return err
		}
		return &IError{
			Errors: []string{"UNKNOWN; error message not parsable"},
		}
	default:
		if err := c.parseError(body); err != nil {
			return err
		}
		return &IError{
			Errors: []string{"UNKNOWN; error message not parsable"},
		}
	}

	return json.Unmarshal(body, result)
}

func (c *IClient) makeSearchHeader() http.Header {
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("App-Id", c.AppID)
	header.Add("Content-Type", "application/json")
	header.Add("Auth-Token", c.Token)
	return header
}

// ApiRequest performs an Iconik API request.
// Args:
// apiPath: The API Resource
// Request The filled request object (eg. SearchCriteriaSchema) to be used.
// Response: The response object (eg. SearchResponse) to be filled out.
func (c *IClient) ApiRequest(apiPath string, request interface{}, response interface{}) error {
	body, err := json.Marshal(request)
	// TODO (@zjames): figuring out header
	header := c.makeSearchHeader()
	if err != nil {
		return err
	}

	if c.Debug {
		log.Println("----")
		log.Printf("apiRequest: %s %s", apiPath, body)
	}

	resp, err := c.authPost(apiPath, bytes.NewReader(body), header)
	if err != nil {
		if c.Debug {
			log.Printf("IClient.post(%s, %v, %v) returned an error: %v\n", apiPath, body, header, err)
		}
		return err
	}

	return c.parseResponse(resp, response)
}
