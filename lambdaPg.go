package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	minio "github.com/minio/minio-go"
)

// Record holds json data from object.
type Record struct {
	Data struct {
		Value struct {
			Records []struct {
				S3 struct {
					Bucket struct {
						Name string `json:"name"`
					} `json:"bucket"`
					Object struct {
						Key string `json:"key"`
					} `json:"object"`
				} `json:"s3"`
			} `json:"Records"`
		} `json:"value"`
	} `json:"data"`
}

func waitForNotification(minioClient *minio.Client, l *pq.Listener) {
	for {
		select {
		case n := <-l.Notify:
			fmt.Println("Received data from channel [", n.Channel, "] :")
			// Prepare notification payload for pretty print
			fmt.Println(n.Extra)
			record := Record{}

			jerr := json.Unmarshal([]byte(n.Extra), &record)
			if jerr != nil {
				fmt.Println("Error processing JSON: ", jerr)
				return
			}

			output.Key = record.Data.Value.Records[0].S3.Bucket.Name
			output.Value = record.Data.Value.Records[0].S3.Object.Key
			output.Parsed = processOCR(minioClient, record.Data.Value.Records[0].S3.Bucket.Name, record.Data.Value.Records[0].S3.Object.Key)

			storeMetaData()
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				fmt.Println("Error processing JSON: ", err)
				return
			}

			output.Metadata = string(prettyJSON.Bytes())
			return
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}

func storeMetaData() {
	_, err := globalPDB.NamedExec(`INSERT INTO bucketmetadata (key,value) VALUES (:key,:value)`,
		map[string]interface{}{
			"key":   output.Value,
			"value": output.Parsed,
		})
	if err != nil {
		fmt.Println("Error inserting metadata: ", err)
		return
	}
}
