package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	iconik "github.com/jzhang919/iconikclient2"
)

// this app will take a local file and upload it to Backblaze B2 and then ingest it into Iconik
func main() {
	appID := flag.String("AppID", "", "Enter your App ID: ")
	token := flag.String("Token", "", "Enter your access token: ")
	debug := flag.Bool("Debug", false, "Debugging")
	fileName := flag.String("Filename", "", "file that you want to upload (local full path)")
	title := flag.String("Title", "", "title you want to see in Iconik")
	collection := flag.String("Collection", "", "collection you want to add the asset to")
	storagePath := flag.String("StoragePath", "/", "storage path you want to save to in B2")
	flag.Parse()

	if *appID == "" || *token == "" || *fileName == "" || *title == "" || *collection == "" {
		log.Fatalf("missing required args: AppID(%s), Token(%s), Filename(%s), Title(%s), Collection(%s)", *appID, *token, *fileName, *title, *collection)
	}
	creds := iconik.Credentials{
		AppID: *appID,
		Token: *token,
	}
	client, err := iconik.NewIClient(creds, "", *debug)
	if err != nil {
		log.Fatalf("Unable to create client: %v\n", err)
	}

	collectionIDs, err := client.GetCollectionIDs(*collection)
	if err != nil {
		log.Fatalf("error getting collectionID: %v", err)
	}
	for i, v := range collectionIDs {
		if i == 0 {
			log.Printf("Using collectionID entry: %v", v)
		} else {
			log.Printf("(unused) collectionID entry: %v", v)
		}
	}
	// open file and get its size, creation date and mimeType
	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("error getting file info: %v", err)
	}
	fileSize := fileInfo.Size()
	// get it's date last modified
	fileDateCreated := fileInfo.ModTime()
	uploadBody, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}
	// get mimetype
	mimeType := http.DetectContentType(uploadBody)

	// create the stubs for the upload
	NAU, err := client.MakeNewAsset(collectionIDs[0].CollectionID, *fileName, *title, *storagePath, mimeType, fileSize, fileDateCreated)
	if err != nil {
		log.Fatalf("error making new asset: %v", err)
	}

	// upload file
	uploadReq, err := http.NewRequest(http.MethodPost,
		NAU.UploadURL,
		bytes.NewBuffer(uploadBody))
	if err != nil {
		log.Fatalf("can't make upload request: %v", err)
	}

	uploadReq.Header.Set("Authorization", NAU.UploadAuthToken)
	uploadReq.Header.Set("X-Bz-File-Name", url.PathEscape(NAU.UploadFilename))
	uploadReq.Header.Set("X-Bz-Content-Sha1", "do_not_verify") // or provide a valid SHA1 checksum
	uploadReq.Header.Add("Content-Type", NAU.MimeType)

	// Send the request
	uploadClient := &http.Client{}
	resp, err := uploadClient.Do(uploadReq)
	if err != nil {
		log.Fatalf("can't upload file: %v", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("can't read upload response: %v", err)
	}
	// Check the response
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Bad status during upload: %s because %s", resp.Status, string(respBody))
	}
	log.Printf("response to B2 upload: %s", string(respBody))

	err = client.FinishUpload(NAU)
	if err != nil {
		log.Fatalf("error is %v", err)
	}

	log.Printf("success!")
}
