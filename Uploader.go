package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/lib/pq"
	"github.com/minio/minio-go"
)

type Profile struct {
	Name    string
	Hobbies []string
}

type Output struct {
	Key    string
	Value  string
	Parsed string
}

var output = Output{}

func main() {
	setUp()
	http.HandleFunc("/", foo)
	http.HandleFunc("/results", bar)
	http.HandleFunc("/upload", upload)
	go listenPG()
	http.ListenAndServe(":3000", nil)

}

func listenPG() {
	minioClient, err := minio.New("192.168.1.15:9000", "minio", "minio123", false)
	if err != nil {
		log.Fatalln(err)
	}
	conninfo := "dbname=minio_events user=postgres password=postgres"

	_, err = sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("watchers")
	if err != nil {
		panic(err)
	}

	fmt.Println("Start monitoring PostgreSQL...")
	for {
		waitForNotification(minioClient, listener)
	}
}
func setUp() {
	minioClient, err := minio.New("192.168.1.15:9000", "minio", "minio123", false)
	if err != nil {
		log.Fatalln(err)
	}

	queueArn := minio.NewArn("minio", "sqs", "", "1", "postgresql")

	queueConfig := minio.NewNotificationConfig(queueArn)
	queueConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)
	//topicConfig.AddFilterPrefix("photos/")
	//topicConfig.AddFilterSuffix(".jpg")

	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(queueConfig)
	err = minioClient.SetBucketNotification("barcodes", bucketNotification)
	if err != nil {
		fmt.Println("Unable to set the bucket notification: ", err)
	}
	fmt.Println("Set Bucket Notification successfully")
	/*// s3Client.TraceOn(os.Stderr)
	// Create a done channel to control 'ListenBucketNotification' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	// Listen for bucket notifications on "mybucket" filtered by prefix, suffix and events.
	for notificationInfo := range minioClient.ListenBucketNotification("barcodes", "", "", []string{
		"s3:ObjectCreated:*",
		"s3:ObjectAccessed:*",
		"s3:ObjectRemoved:*",
	}, doneCh) {
		if notificationInfo.Err != nil {
			log.Println(notificationInfo.Err)
		} else {
			log.Println("******************")
			log.Println(notificationInfo)
			log.Println("******************")
		}
	}*/
}
func bar(w http.ResponseWriter, r *http.Request) {
	js, err := json.Marshal(output)
	if err != nil {
		fmt.Print(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func foo(w http.ResponseWriter, r *http.Request) {
	profile := Profile{"Alex", []string{"snowboarding", "programming"}}

	lp := path.Join("templates", "layout.html")
	fp := path.Join("templates", "index.html")
	fmt.Println("In foo")
	// Note that the layout file must be the first parameter in ParseFiles
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling upload handler")
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fname := header.Filename
	fmt.Printf("Filename %s\n", fname)

	minioClient, err := minio.New("192.168.1.15:9000", "minio", "minio123", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("Calling FPutObject")

	n, err := minioClient.PutObject("barcodes", fname, file, -1, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Successfully put object %d\n", n)
}
