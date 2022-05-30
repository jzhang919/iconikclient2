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
	// fmt.Print("Enter your App ID: ")
	// var app_id string;
	// fmt.Scanln(&app_id)

	// fmt.Print("Enter your access token: ")
	// var token string;
	// fmt.Scanln(&token)

	appIDFlag := flag.String("AppID", "", "Enter your App ID: ");
	tokenFlag := flag.String("Token", "", "Enter your access token: ");
	// debugFlag := flag.Bool("DebugFlag", false, "Debugging")

	flag.Parse()

	// fmt.Printf("App ID Value: %s", *appIDFlag)
	// fmt.Printf("Token Value: %s", *tokenFlag)


	creds := Credentials{
		AppID: *appIDFlag,
		Token: *tokenFlag,
	}
	client, _ := NewIClient(creds, "")
	resp, _ := client.SearchWithTag("AYMLeadershipCamp")
	//resp, _ := client.SearchWithTitle("aym_onboarding_intro.mp4")
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