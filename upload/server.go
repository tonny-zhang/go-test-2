package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"time"
)

const (
	// UPLOADDIR UPLOADDIR
	UPLOADDIR = "./upload"
	// KEY KEY
	KEY = "hello"
)

func toMd5(msg string) (msgMd5 string) {
	md5Inst := md5.New()
	md5Inst.Write([]byte(msg))
	return hex.EncodeToString(md5Inst.Sum([]byte(nil)))
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			if e, ok := recover().(error); ok {
				http.Error(rw, e.Error(), http.StatusInternalServerError)
				log.Printf("WARN: panic in %v - %v", fn, e)
				log.Println(string(debug.Stack()))
			}
		}()

		fn(rw, req)
	}
}
func uploadHandle(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		file, header, err := req.FormFile("file_upload")
		checkErr(err)

		defer file.Close()
		filename := header.Filename

		filenameNew := toMd5(time.Now().String()) + path.Ext(filename)

		os.MkdirAll(UPLOADDIR, os.ModeDir)
		filepath := UPLOADDIR + "/" + filenameNew
		target, err := os.Create(filepath)
		checkErr(err)
		defer target.Close()

		_, err1 := io.Copy(target, file)
		checkErr(err1)

		fmt.Fprintf(rw, "any, %q", html.EscapeString(filepath))
	} else if req.Method == "GET" {
		tmpl := `<!doctype html>
			<html>

			<head>
				<meta charset="utf-8">
				<title>List</title>
			</head>

			<body>
				<form action="http://localhost:8000/upload" method="POST" enctype="multipart/form-data">
					<input type="file" name="file_upload" />
					<input type="submit" value="upload" />
				</form>
			</body>

			</html>`

		rw.Write([]byte(tmpl))
	}
}

func main() {
	http.HandleFunc("/upload", safeHandler(uploadHandle))

	log.Fatal(http.ListenAndServe(":8000", nil))
}
