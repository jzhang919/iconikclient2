package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	iconik "github.com/jzhang919/iconikclient2"
)

type AssetProxy struct {
	assetID string
	proxyID string
}

func main() {
	appIDFlag := flag.String("AppID", "", "Enter your App ID: ")
	tokenFlag := flag.String("Token", "", "Enter your access token: ")
	debugFlag := flag.Bool("Debug", false, "Debugging")
	iconikID := flag.String("iconikID", "", "ID to get signed URL for")
	flag.Parse()

	creds := iconik.Credentials{
		AppID: *appIDFlag,
		Token: *tokenFlag,
	}
	client, _ := iconik.NewIClient(creds, "", *debugFlag)

	url, err := client.GenerateSignedFileUrl(*iconikID)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("Signed URL: %s\n", url)
	// go and GET the url
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to perform GET request: %v", err)
	}
	defer resp.Body.Close()

	// Print the status code for reference
	fmt.Printf("Status Code: %d\n", resp.StatusCode)

	// Iterate through and print all the headers
	fmt.Println("Response Headers:")
	for key, value := range resp.Header {
		fmt.Printf("%s: %s\n", key, value)
	}

}
