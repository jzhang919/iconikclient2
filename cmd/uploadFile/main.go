package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
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

	// check if we are doing single file upload or multipart
	if NAU.MultipartFileID != "" {
		partNum := 1
		shas := []string{}

		for {
			startPos := int64(partNum-1) * iconik.MULTIPART_FILESIZE_THRESHOLD
			endPos := startPos + iconik.MULTIPART_FILESIZE_THRESHOLD
			if endPos > fileSize {
				endPos = fileSize
			}
			if startPos >= fileSize {
				break
			}

			uploadReq, err := http.NewRequest(http.MethodPost,
				NAU.UploadURL,
				bytes.NewBuffer(uploadBody[startPos:endPos]))
			if err != nil {
				log.Fatalf("can't make upload request: %v", err)
			}

			// Calculate SHA1 of the body slice
			hasher := sha1.New()
			hasher.Write(uploadBody[startPos:endPos])
			sha1Hash := hasher.Sum(nil)

			uploadReq.Header.Set("Authorization", NAU.UploadAuthToken)
			uploadReq.Header.Set("X-Bz-Part-Number", fmt.Sprintf("%d", partNum))
			uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", (endPos-startPos)))
			uploadReq.Header.Set("X-Bz-Content-Sha1", fmt.Sprintf("%x", sha1Hash))

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
			type MultiResp struct {
				ContentSha1 string `json:"contentSha1"`
			}
			var multiResp MultiResp
			err = json.Unmarshal(respBody, &multiResp)
			if err != nil {
				log.Fatalf("error unmarshalling multipart response: %v", err)
			}
			shas = append(shas, multiResp.ContentSha1)
			partNum++
		}
		NAU.Sha1List = shas
	} else {
		// upload file
		uploadReq, err := http.NewRequest(http.MethodPost,
			NAU.UploadURL,
			bytes.NewBuffer(uploadBody))
		if err != nil {
			log.Fatalf("can't make upload request: %v", err)
		}

		// Calculate SHA1 of the body slice
		hasher := sha1.New()
		hasher.Write(uploadBody)
		sha1Hash := hasher.Sum(nil)

		uploadReq.Header.Set("Authorization", NAU.UploadAuthToken)
		uploadReq.Header.Set("X-Bz-File-Name", url.PathEscape(NAU.UploadFilename))
		uploadReq.Header.Set("X-Bz-Content-Sha1", fmt.Sprintf("%x", sha1Hash))
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
	}

	err = client.FinishUpload(NAU)
	if err != nil {
		log.Fatalf("error is %v", err)
	}

	log.Printf("success!")
}
