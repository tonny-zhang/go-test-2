package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	targetURL := "http://127.0.0.1:8000/upload"

	filename := "../upload.html"
	fileWriter, err := bodyWriter.CreateFormFile("file_upload", filename)

	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println("error opening file")
		return
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	resp, err := http.Post(targetURL, contentType, bodyBuf)
	fmt.Println(bodyBuf)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	fmt.Println(resp.Status)
	fmt.Println(string(respBody))
	return
}
