package iconik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendRequest(t *testing.T) {
	expected := "abc"
	returnStruct := SearchResponse{Objects: []IconikObject{{Id: expected, Files: []IconikFile{{Name: "test"}}}}}

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if strings.TrimPrefix(req.URL.String(), "/") == searchEndpoint {
			payload, _ := json.Marshal(returnStruct)
			rw.Write(payload)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, server.URL, false)

	req := makeSearchBody("testTitle", "testTag")
	resp, err := client.SearchWithTag("Teaching")
	if err != nil {
		t.Fatalf("ApiRequest(%s, %v) failed: %v", searchEndpoint, req, err)
	}
	if len(resp.Objects) != len(returnStruct.Objects) {
		t.Errorf("Got %d objects in response; wanted %d objects", len(resp.Objects), len(returnStruct.Objects))
	}
}

func TestIClient_GenerateSignedProxyUrl(t *testing.T) {
	expected := "https://test.com/url"
	returnStruct := GetResponse{Objects: []Object{{URL: expected}}}
	assetId := "testAssetId"

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if strings.TrimPrefix(req.URL.String(), "/") == fmt.Sprintf(proxyEndpointTemplate, assetId) {
			payload, _ := json.Marshal(returnStruct)
			rw.Write(payload)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, server.URL, false)
	url, err := client.GenerateSignedProxyUrl(assetId)
	if err != nil {
		t.Fatalf("GenerateSignedProxyUrl(%s got %v; wanted no error)", assetId, err)
	}
	if url != expected {
		t.Errorf("GenerateSignedProxyUrl(%s) got %s; wanted %s", assetId, url, expected)
	}
}

func TestIClient_GenerateSignedFileUrl(t *testing.T) {
	expected := "https://test.com/url"
	returnStruct := GetResponse{Objects: []Object{{URL: expected}}}
	assetId := "testAssetId"

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if strings.TrimPrefix(req.URL.String(), "/") == fmt.Sprintf(fileEndpointTemplate, assetId) {
			payload, _ := json.Marshal(returnStruct)
			rw.Write(payload)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, server.URL, false)
	url, err := client.GenerateSignedFileUrl(assetId)
	if err != nil {
		t.Fatalf("GenerateSignedFileUrl(%s got %v; wanted no error)", assetId, err)
	}
	if url != expected {
		t.Errorf("GenerateSignedFileUrl(%s) got %s; wanted %s", assetId, url, expected)
	}
}

func TestIClient_GenerateKeyframeUrl(t *testing.T) {
	expected := "https://test.com/url"
	returnStruct := GetResponse{Objects: []Object{{URL: expected, Type: "KEYFRAME"}}}
	assetId := "testAssetId"

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if strings.TrimPrefix(req.URL.String(), "/") == fmt.Sprintf(keyframeEndpointTemplate, assetId) {
			payload, _ := json.Marshal(returnStruct)
			rw.Write(payload)
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	creds := Credentials{
		AppID: "testAppID",
		Token: "testToken",
	}
	client, _ := NewIClient(creds, server.URL, false)
	url, err := client.GetKeyframeUrl(assetId)
	if err != nil {
		t.Fatalf("GetKeyframeUrl(%s got %v; wanted no error)", assetId, err)
	}
	if url != expected {
		t.Errorf("GetKeyframeUrl(%s) got %s; wanted %s", assetId, url, expected)
	}
}
