package main

import (
	sysZip "archive/zip"
	"io"
	"log"
	"os"
	"strings"

	"github.com/alexmullins/zip"
)

func createZip1(filename string, files []map[string]string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	// contents := []byte("Hello World")

	// write a password zip
	// raw := new(bytes.Buffer)
	// zipw := zip.NewWriter(raw)
	// w, err := zipw.Encrypt("hello.txt", "golang")

	zipWriter := zip.NewWriter(newZipFile)

	// w, err := zipWriter.Encrypt("hello.txt", "golang")
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip1(zipWriter, file["path"], file["name"]); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip1(zipWriter *zip.Writer, filename string, filenameInZip string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := sysZip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	// header.Name = filename
	filenameInZip = strings.Join(strings.Split(filenameInZip, "\\"), "/")
	header.Name = filenameInZip

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func createZip(filename string, files []map[string]string) error {
	fzip, err := os.Create(`./test.zip`)
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer fzip.Close()

	zipw := zip.NewWriter(fzip)
	defer zipw.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipw, file["path"], file["name"], file["pwd"]); err != nil {
			return err
		}
	}

	zipw.Flush()
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string, filenameInZip string, pwd string) error {
	w, err := zipWriter.Encrypt(filenameInZip, pwd)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()
	_, err = io.Copy(w, fileToZip)
	return err
}
func main() {
	files := make([]map[string]string, 0)
	files = append(files, map[string]string{
		"path": "/Users/tonny/source/doc/xy3/每日充值.xls",
		"name": "1.xls",
		"pwd":  "a",
	})
	files = append(files, map[string]string{
		"path": "/Users/tonny/source/doc/xy3/活跃度奖励.xlsx",
		"name": "2.xls",
		"pwd":  "b",
	})

	createZip("./test.zip", files)

	// contents := []byte("Hello World")
	// fzip, err := os.Create(`./test.zip`)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// zipw := zip.NewWriter(fzip)
	// defer zipw.Close()
	// w, err := zipw.Encrypt(`test.txt`, `golang`)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = io.Copy(w, bytes.NewReader(contents))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// zipw.Flush()
}
