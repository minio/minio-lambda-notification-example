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

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/minio/minio-go"
)

// Output - holds JSON Output
type Output struct {
	Key      string
	Value    string
	Parsed   string
	Metadata string
}

// This is the const that holds tricord postgres connection info.
const (
	DBUSER     = "postgres"
	DBPASSWORD = "postgres"
	DBNAME     = "minio_events"
)

var output = Output{}
var globalPDB *sqlx.DB

func main() {
	setUp()
	http.HandleFunc("/", foo)
	http.HandleFunc("/results", bar)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/search", searchMetaData)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	go pgHook()
	http.ListenAndServe(":3000", nil)
}

func pgHook() {
	minioClient, err := minio.New("192.168.1.118:9000", "minio", "minio123", false)
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
	for {
		waitForNotification(minioClient, listener)
	}
}

func initDB() {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=verify-ca host=localhost", DBUSER, DBPASSWORD, DBNAME)
	pdb, err := sqlx.Connect("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}

	// Set new postgres instance.
	globalPDB = pdb
}

func setUp() {
	initDB()
	minioClient, err := minio.New("192.168.1.118:9000", "minio", "minio123", false)
	if err != nil {
		log.Fatalln(err)
	}
	queueArn := minio.NewArn("minio", "sqs", "", "1", "postgresql")
	queueConfig := minio.NewNotificationConfig(queueArn)
	queueConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)
	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(queueConfig)
	err = minioClient.SetBucketNotification("barcodes", bucketNotification)
	if err != nil {
		fmt.Println("Unable to set the bucket notification: ", err)
	}
	fmt.Println("Set Bucket Notification successfully")
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
	lp := path.Join("templates", "layout.html")
	fp := path.Join("templates", "index.html")
	// Note that the layout file must be the first parameter in ParseFiles
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, ""); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fname := header.Filename

	minioClient, err := minio.New("192.168.1.118:9000", "minio", "minio123", false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	n, err := minioClient.PutObject("barcodes", fname, file, header.Size, minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Successfully put object %d\n", n)
}

func searchMetaData(w http.ResponseWriter, req *http.Request) {

	mdt := []Output{}

	var keyTemp, valueTemp string
	var err error

	rows, err := globalPDB.Queryx(`select * from bucketmetadata`)

	if err != nil {

		return
	}
	for rows.Next() {
		err = rows.Scan(&keyTemp, &valueTemp)
		if err != nil {
			fmt.Println("minioTable query scan error:", err)
			continue
		}
		//construct & populate a MinioDataTable struct instance.
		mymdt := Output{
			keyTemp,
			valueTemp,
			"n/a",
			"n/a",
		}
		mdt = append(mdt, mymdt)
	}
	js, err := json.Marshal(mdt)
	if err != nil {
		fmt.Println("Unable to marshall data:", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
