package iconik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSendRequest(t *testing.T) {
	expected := SearchResponse{Objects: []IconikObject{{Id: "abc", Files: []IconikFile{IconikFile{Name: "test"}}}}}

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.String() == iconikHost+searchEndpoint {
			payload, _ := json.Marshal(expected)
			rw.Write(payload)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, "")

	req := makeSearchBody("testTag")
	resp, err := client.SearchWithTag("Teaching")
	if err != nil {
		t.Fatalf("ApiRequest(%s, %v) failed: %v", searchEndpoint, req, err)
	}
	if len(resp.Objects) != len(expected.Objects) {
		t.Errorf("Got %d objects in response; wanted %d objects", len(resp.Objects), len(expected.Objects))
	}
}

func TestIClient_GenerateSignedProxyUrl(t *testing.T) {
	expected := "https://test.com/url"
	assetId := "testAssetId"
	proxyId := "testProxyId"

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.String() == iconikHost+fmt.Sprintf(proxyEndpointTemplate, assetId, proxyId) {
			payload, _ := json.Marshal(expected)
			rw.Write(payload)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, "")
	url, err := client.GenerateSignedProxyUrl(assetId, proxyId)
	if err != nil {
		t.Fatalf("GenerateSignedProxyUrl(%s, %s) got %v; wanted no error)", assetId, proxyId, err)
	}
	if url != expected {
		t.Errorf("GenerateSignedProxyUrl(%s, %s) got %s; wanted %s", assetId, proxyId, url, expected)
	}
}
