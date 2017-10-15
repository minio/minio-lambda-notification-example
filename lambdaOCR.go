package main

import (
	"fmt"
	"io"
	"os"

	minio "github.com/minio/minio-go"
	"github.com/otiai10/gosseract"
)

func processOCR(minioClient *minio.Client, bucketname string, objectname string) string {
	// This is the simplest way :)

	object, err := minioClient.GetObject(bucketname, objectname, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return ""
	}
	localFile, err := os.Create("/tmp/" + objectname)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	if _, err = io.Copy(localFile, object); err != nil {
		fmt.Println(err)
		return ""
	}
	// Using client
	fmt.Println("I am here")
	client, _ := gosseract.NewClient()
	out, _ := client.Src(localFile.Name()).Out()
	//fmt.Println(client)
	//fmt.Println(out)
	return out
}
