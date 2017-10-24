package main

import (
	"fmt"
	"image/jpeg"
	"io"
	"os"

	barcode "github.com/bieber/barcode"
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
	client, _ := gosseract.NewClient()
	out, _ := client.Src(localFile.Name()).Out()

	fin, _ := os.Open(localFile.Name())
	defer fin.Close()
	src, _ := jpeg.Decode(fin)
	img := barcode.NewImage(src)
	scanner := barcode.NewScanner().SetEnabledAll(true)

	symbols, _ := scanner.ScanImage(img)
	for _, s := range symbols {
		fmt.Println(s.Type.Name(), s.Data, s.Quality, s.Boundary)
		out += s.Data
	}
	output.Parsed = out
	return out
}
