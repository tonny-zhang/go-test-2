package main

import (
	"archive/zip"
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"test/packageAndUpload/config"
	"test/packageAndUpload/fastwalk"
	"test/packageAndUpload/uploader"
	"time"
)

var step = 1

// 得到一个步骤编号
func getStep(isNext bool) int {
	if isNext {
		step++
	}
	return step
}
func printErrAndWaitExit(err interface{}) {
	log.Fatal(err)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n\n回车退出...")
	reader.ReadByte()
	os.Exit(0)
}
func walkDir(dirPath string, excludeExt map[string]bool) (files []string, err error) {
	files = make([]string, 0, 30)

	var mu sync.Mutex
	err = fastwalk.Walk(dirPath, func(filename string, t os.FileMode) error {
		mu.Lock()
		defer mu.Unlock()
		if !t.IsDir() {
			// filenameR, _ := filepath.Rel(dirPath, filename)
			ext := filepath.Ext(filename)
			if _, ok := excludeExt[ext]; !ok {
				// 过滤.目录（即隐藏目录）
				if b, _ := regexp.MatchString(`/\.\w`, filename); !b {
					files = append(files, filepath.ToSlash(filename))
				}

			}
		}
		return nil
	})
	return files, err
}

func getLastVersion(dirPath string) (vBig, vSmall int) {
	dirs, err := ioutil.ReadDir(dirPath)

	versionBig := 0
	versionSmall := 0
	if err == nil {
		for _, dir := range dirs {
			if dir.IsDir() {
				dirName := dir.Name()
				b, _ := regexp.MatchString(`\d+\.\d+`, dirName)
				if b {
					arr := strings.Split(dirName, ".")
					if len(arr) == 2 {
						vBig, _ := strconv.Atoi(arr[0])
						vSmall, _ := strconv.Atoi(arr[1])
						if vBig > versionBig {
							versionBig = vBig
							versionSmall = vSmall
						} else if vSmall > versionSmall {
							versionSmall = vSmall
						}
					}
				}
			}
		}
	}

	return versionBig, versionSmall
}
func getFileMd5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
func isFileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
func createZip(filename string, files []map[string]string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file["path"], file["name"]); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string, filenameInZip string) error {

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

	header, err := zip.FileInfoHeader(info)
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

type resultContent struct {
	Files        map[string]string
	VersionGig   int
	VersionSmall int
}

func deal(confDir string) {
	fmt.Println("开始处理...")

	tReadConf := time.Now()
	fmt.Printf("%d 准备读取配置文件\n", getStep(false))

	var conf config.ConfLocal
	confContent, _ := ioutil.ReadFile(confDir)
	err := json.Unmarshal(confContent, &conf)
	if err != nil {
		printErrAndWaitExit(err)
	}

	fmt.Printf("%d 读取配置文件完成, 用时%v\n", getStep(false), time.Since(tReadConf))
	dirPath := conf.DirLocal
	dirPathTmp := conf.DirLocalTmp
	extExclude := conf.ExtExclude

	if !isFileExists(dirPath) {
		printErrAndWaitExit(fmt.Sprintf("[%s]不存在", dirPath))
	}
	if !isFileExists(dirPathTmp) {
		printErrAndWaitExit(fmt.Sprintf("[%s]不存在", dirPathTmp))
	}

	dirPathTmpFiles := filepath.Join(dirPathTmp, "tmp")

	os.MkdirAll(dirPathTmp, 0777)
	os.Mkdir(dirPathTmpFiles, 0777)

	vBig, vSmall := getLastVersion(dirPathTmp)

	dirPathVersion := strconv.Itoa(vBig) + "." + strconv.Itoa(vSmall)
	pathResultFile := filepath.Join(dirPathTmp, dirPathVersion, "result.json")

	tReadResultPrev := time.Now()
	fmt.Printf("%d 准备读取上一个版本[%s]的结果\n", getStep(true), dirPathVersion)
	resultPrev := resultContent{}
	if isFileExists(pathResultFile) {
		jsonContent, _ := ioutil.ReadFile(pathResultFile)
		json.Unmarshal(jsonContent, &resultPrev)
	} else {
		fmt.Printf("[%s]不存在\n", pathResultFile)
	}
	fmt.Printf("%d 读取上一个版本的结果完成, 用时%v\n", getStep(false), time.Since(tReadResultPrev))

	extExcludeMap := make(map[string]bool)
	for _, key := range extExclude {
		extExcludeMap[key] = true
	}

	tGetFiles := time.Now()
	fmt.Printf("%d 遍历目录[%s]\n", getStep(true), dirPath)
	files, _ := walkDir(dirPath, extExcludeMap)
	lenTotal := len(files)
	fmt.Printf("%d 遍历目录完成 用时 %v, 共有[%d]个文件要处理\n", getStep(false), time.Since(tGetFiles), lenTotal)

	filesNew := make([]map[string]string, 0, 30)
	resultFiles := make(map[string]string)

	tDealFiles := time.Now()
	fmt.Printf("%d 开始处理文件\n", getStep(true))
	for i, file := range files {
		md5OfFile, _ := getFileMd5(file)
		if md5Old, ok := resultPrev.Files[file]; !ok || md5Old != md5OfFile {
			filePathRel, _ := filepath.Rel(dirPath, file)
			fileNew := filepath.Join(dirPathTmpFiles, filePathRel)
			os.MkdirAll(filepath.Dir(fileNew), 0777)
			copy(file, fileNew)
			filesNew = append(filesNew, map[string]string{
				"path": filepath.ToSlash(fileNew),
				"name": filepath.ToSlash(filePathRel),
			})
			fmt.Printf("%d/%d Y [%s] [%s] [%s]\n", i, lenTotal, file, md5OfFile, md5Old)
		} else {
			fmt.Printf("%d/%d N [%s]\n", i, lenTotal, file)
		}
		resultFiles[file] = md5OfFile
	}
	resultPrev.Files = resultFiles

	lenFilesNew := len(filesNew)
	fmt.Printf("%d 处理文件完成，共更新[%d]个文件，用时[%v]\n", getStep(false), lenFilesNew, time.Since(tDealFiles))

	for i, file := range filesNew {
		fmt.Printf("更改文件 %d/%d %s\n", i, lenFilesNew, file["name"])
	}

	if lenFilesNew > 0 {
		tCreateResult := time.Now()
		fmt.Printf("%d 准备生成结果文件\n", getStep(true))
		if resultPrev.VersionGig != conf.VersionGig {
			resultPrev.VersionGig = conf.VersionGig
			resultPrev.VersionSmall = 0
		} else {
			resultPrev.VersionSmall++
		}
		resultStr, _ := json.Marshal(resultPrev)

		versionCurrent := fmt.Sprintf("%d.%d", resultPrev.VersionGig, resultPrev.VersionSmall)

		pathVersionCurrent := path.Join(dirPathTmp, versionCurrent)
		os.MkdirAll(pathVersionCurrent, 0777)

		pathResultFileNextVersion := filepath.Join(pathVersionCurrent, "result.json")
		ioutil.WriteFile(pathResultFileNextVersion, resultStr, 0777)
		fmt.Printf("%d 生成结果文件 [%s] 用时%v\n", getStep(false), pathResultFileNextVersion, time.Since(tCreateResult))

		tCreateZip := time.Now()
		fmt.Printf("%d 准备生成zip文件\n", getStep(true))

		// zipName := time.Now().Format("20060102150405.zip")
		zipName := versionCurrent + ".zip"
		zipPath := filepath.Join(pathVersionCurrent, zipName)

		createZip(zipPath, filesNew)

		fmt.Printf("%d 生成zip文件 [%s], 用时%v\n", getStep(false), zipPath, time.Since(tCreateZip))

		tCreateUpdateFile := time.Now()
		fmt.Printf("%d 准备生成更新文件\n", getStep(true))
		var jsonUpdateFile = make(map[string][]map[string]string)
		pathUpdateFilePrev := path.Join(dirPathTmp, dirPathVersion, "version.txt")
		pathUpdateFile := path.Join(pathVersionCurrent, "version.txt")

		var dataUpdateFile = make([]map[string]string, 0, 0)
		if isFileExists(pathUpdateFilePrev) {
			jsonContent, _ := ioutil.ReadFile(pathUpdateFilePrev)
			json.Unmarshal(jsonContent, &jsonUpdateFile)
			dataUpdateFile = jsonUpdateFile["data"]
		}

		var item = map[string]string{
			"version": versionCurrent,
			"file":    zipName,
		}
		jsonUpdateFile["data"] = append(dataUpdateFile, item)

		strUpdateFile, _ := json.Marshal(jsonUpdateFile)
		ioutil.WriteFile(pathUpdateFile, strUpdateFile, 0777)
		fmt.Printf("%d 生成更新文件[%s], 用时%v\n", getStep(false), pathUpdateFile, time.Since(tCreateUpdateFile))

		// 删除生成的临时文件夹
		os.RemoveAll(dirPathTmpFiles)
		fmt.Print("\n是否上传更新zip, Y/N?\n")

		reader := bufio.NewReader(os.Stdin)
		b, _ := reader.ReadByte()
		if b != 'Y' && b != 'y' {
			os.Exit(0)
		}

		tUpload := time.Now()
		fmt.Printf("%d 准备上传zip文件\n", getStep(true))
		upload(zipPath, conf)
		fmt.Printf("%d 上传zip文件用时%v\n", getStep(false), time.Since(tUpload))

		tUploadUpdateFile := time.Now()
		fmt.Printf("%d 准备上传更新文件\n", getStep(true))
		upload(pathUpdateFile, conf)
		fmt.Printf("%d 上传更新文件完成，用时%v\n", getStep(false), time.Since(tUploadUpdateFile))
	} else {
		fmt.Println("没有要处理的更新文件")
		// 删除生成的临时文件夹
		os.RemoveAll(dirPathTmpFiles)
	}
}

func upload(sourceFile string, conf config.ConfLocal) {
	if conf.QnKey != "" && conf.QnSecret != "" && conf.QnBucket != "" {
		uploader.UploadToQN(conf, sourceFile)
	}
}

// {
//     "DirLocal": "本地路径",
// 	   "DirLocalTmp": "本地生成路径",
// 	   "ExtExclude": [".asset", ".meta"],
//     "Host": "183.60.204.134",
//     "Port": 22,
//     "User": "用户名",
//     "Pwd": "密码",
//     "DirRemote": "远程路径",
//     "VersionGig": 1
// }
func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		printErrAndWaitExit(err)
	} else {
		confDir := path.Join(dir, "conf.json")
		// fmt.Println(dir)
		// confDir = "/Users/tonny/source/go/src/test/packageAndUpload/conf.json"
		if isFileExists(confDir) {
			deal(confDir)
		} else {
			fmt.Printf("[%s]不存在", confDir)
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n\n回车退出...")
	reader.ReadByte()
	os.Exit(0)

	// dirPath := "/Users/tonny/source/test/tolua"
	// deal(dirPath)
}

// func main() {
// 	dirPath := "/Users/tonny/source/test/tolua"
// 	fastwalk.Walk(dirPath, func(filename string, t os.FileMode) error {
// 		fmt.Println(filename, t.IsDir())
// 		return nil
// 	})
// }
