package main

import (
	"flag"
	"fmt"
)

type AssetProxy struct {
	assetID string
	proxyID string
}

func main() {

	appID := flag.String("appid", "Please enter your App Id:", "Application Identification")
	flag.Parse()

	tokenValue := flag.String("tokenValue", "Please enter your Token: ", "Token")
	flag.Parse()

	creds := Credentials{
		AppID: *appID,
		Token: *tokenValue,
	}

	client, _ := NewIClient(creds, "")
	resp, _ := client.SearchWithTag("GPTeaching")
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
