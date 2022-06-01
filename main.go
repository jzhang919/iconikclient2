package main

import (
	"fmt"
	"flag"
)

type AssetProxy struct {
	assetID string
	proxyID string
}

func main() {
	appIDFlag := flag.String("AppID", "", "Enter your App ID: ");
	tokenFlag := flag.String("Token", "", "Enter your access token: ");
	debugFlag := flag.Bool("Debug", false, "Debugging")
	flag.Parse()

	creds := Credentials{
		AppID: *appIDFlag,
		Token: *tokenFlag,
	}
	client, _ := NewIClient(creds, "", *debugFlag)
	resp, _ := client.SearchWithTitleAndTag("aym_onboarding_intro", "")

	ids := []AssetProxy{}
	for _, objects := range resp.Objects {
		for _, proxy := range objects.Proxies {
			ids = append(ids, AssetProxy{objects.Id, proxy.Id})
		}
		for _, file := range objects.Files {
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