package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSendRequest(t *testing.T) {
	expected := SearchResponse{Objects: []IconikObject{{[]IconikFile{IconikFile{Name: "test"}}}}}
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.String() == searchUrl {
			payload, _ := json.Marshal(expected)
			rw.Write(payload)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	i := IconikClient{client: server.Client()}
	header := makeSearchHeader()
	body, err := makeSearchBody(tag)
	if err != nil {
		t.Fatalf("makeSearchBody() failed: %v", err)
	}
	req, err := makeSearchRequest(searchUrl, header, body)
	if err != nil {
		t.Fatalf("makeSearchRequest(%s, %v, %v) failed: %v", searchUrl, header, body, err)
	}
	searchResponse := i.sendRequest(req)
	if len(searchResponse.Objects) != len(expected.Objects) {
		t.Errorf("Got %d objects in response; wanted %d objects", len(searchResponse.Objects), len(expected.Objects))
	}
}
