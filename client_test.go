package iconik

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testSendRequest(t *testing.T) {
	expected := SearchResponse{Objects: []IconikObject{{[]IconikFile{IconikFile{Name: "test"}}}}}

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
