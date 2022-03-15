package iconik

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func generateSearchBody(tag string) SearchCriteriaSchema {
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

func testSendRequest(t *testing.T) {
	searchEndpoint := "search/v1/search/"
	expected := SearchResponse{Objects: []IconikObject{{[]IconikFile{IconikFile{Name: "test"}}}}}

	//Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.String() == iconikHost + searchEndpoint {
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

	req := generateSearchBody("testTag")
	resp := &SearchResponse{}
	err := client.ApiRequest(searchEndpoint, req, resp)
	if err != nil {
		t.Fatalf("ApiRequest(%s, %v) failed: %v", searchEndpoint, req, err)
	}
	if len(resp.Objects) != len(expected.Objects) {
		t.Errorf("Got %d objects in response; wanted %d objects", len(resp.Objects), len(expected.Objects))
	}
}
