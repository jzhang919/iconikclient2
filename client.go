package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/url"

	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	tag = "GPTeaching"
)

const (
	searchUrl = "https://app.iconik.io/API/search/v1/search/"
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

// Iconik API Client

type IconikClient struct {
	client *http.Client
}

func (i *IconikClient) sendRequest(r *http.Request) SearchResponse {
	response, err := i.client.Do(r)
	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject SearchResponse
	json.Unmarshal(responseData, &responseObject)
	return responseObject
}

func makeSearchHeader() http.Header {
	header := make(http.Header)
	header.Add("accept", "application/json")
	header.Add("App-Id", "bb4bb090-a100-11ec-8fe5-dac11cea8e63")
	header.Add("Content-Type", "application/json")
	header.Add("Auth-Token", "eyJhbGciOiJIUzI1NiIsImlhdCI6MTY0Njk3ODUyMCwiZXhwIjoxOTYyNDM4NTIwfQ.eyJpZCI6ImM0MmU4ZTA4LWExMDAtMTFlYy1hYTUwLTM2ZjFkMmVmNjY1YyJ9.Am9iOHu2FOpsMME3QiEWJv_xk4sCJYaRAjJvAaATIC8")
	header.Add("generated_signed_url", "true")
	return header
}

func makeSearchBody(tag string) ([]byte, error) {
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
	body, err := json.Marshal(schema)
	return body, err
}

func makeSearchRequest(apiUrl string, header http.Header, body []byte) (*http.Request, error) {
	url, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL:    url,
		Header: header,
		Body:   ioutil.NopCloser(bytes.NewReader(body)),
	}
	return req, nil
}

func main() {
	flag.StringVar(&tag, "tag", "GPTeaching", "Tag to search for all GPTeaching assets")
	client := IconikClient{client: &http.Client{}}
	header := makeSearchHeader()
	body, err := makeSearchBody(tag)
	if err != nil {
		fmt.Errorf("makeSearchBody() failed: %v", err)
		return
	}
	req, err := makeSearchRequest(searchUrl, header, body)
	if err != nil {
		fmt.Errorf("makeSearchRequest(%s, %v, %v) failed: %v", searchUrl, header, body, err)
	}
	responseObject := client.sendRequest(req)

	// TODO(zjames@): Replace printing matching files' names with getting their presigned URLs.
	for _, objects := range responseObject.Objects {
		for _, file := range objects.Files {
			fmt.Println(file.Name)
		}
	}
}
