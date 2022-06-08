package main

import (
	"fmt"
	"flag"
)

const (
	testAssetTitle	= "Asset Title"
	testDirectory	= "Churches/_Churchwide/Video Footage Archive/2019/Berk_AYM/"
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

	// Upload File
	assetID, _ := client.GenerateAssetID(testAssetTitle);
	formatID, _ := client.GenerateFormatID(assetID, "format1")
	testVideoComponent, _ := client.GenerateComponentID(assetID, formatID, "videoComponent", "VIDEO")
	testAudioComponent, _ := client.GenerateComponentID(assetID, formatID, "videoComponent", "AUDIO")
	components := []string{testVideoComponent, testAudioComponent}
	fileSetID, _ := client.GenerateFileSetID(assetID, formatID, testDirectory, components, "fileset1")
	tags := []string{"tag1", "tag2"}
	
	// New File Created
	newFile, _ := client.CreateFile(assetID, testDirectory, fileSetID, "New File Name", "FILE", tags)
	fmt.Printf("Asset '%s' created. With new file name: %s", testAssetTitle, newFile.Name)

	// Search for New File
	resp, _ := client.SearchWithTitleAndTag(testAssetTitle, "")

	for _, objects := range resp.Objects {
		for _, file := range objects.Files {
			fmt.Println(file.Name)
		}
	}



}