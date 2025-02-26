package main

import (
	"flag"
	"fmt"

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
	searchTitle := flag.String("title", "", "Title you are searching for")
	searchTag := flag.String("", "", "Tags you are searching for")
	flag.Parse()

	creds := iconik.Credentials{
		AppID: *appIDFlag,
		Token: *tokenFlag,
	}
	client, _ := iconik.NewIClient(creds, "", *debugFlag)
	resp, _ := client.SearchWithTitleAndTag(*searchTitle, *searchTag, false)

	ids := []AssetProxy{}
	for _, object := range resp.Objects {
		for _, proxy := range object.Proxies {
			ids = append(ids, AssetProxy{object.Id, proxy.Id})
		}
		for _, file := range object.Files {
			fmt.Println(file.Name)
		}
	}
	for _, id := range ids {
		url, err := client.GenerateSignedProxyUrl(id.assetID)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println(url)
	}

}
