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
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func walkDir(dirPath string) (files []string, err error) {
	files = make([]string, 0, 30)

	err = filepath.Walk(dirPath, func(filename string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			// filenameR, _ := filepath.Rel(dirPath, filename)
			files = append(files, filename)
		}
		return nil
	})

	return files, err
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

type serverConfig struct {
	host     string
	port     int
	username string
	passwd   string
	path     string
}

func upload(sourceFile string, conf serverConfig) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// defer sshClient.Close()
	// defer sftpClient.Close()
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(conf.passwd))
	// auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            conf.username,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", conf.host, conf.port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err == nil {
		// defer sshClient.Close()
		if sftpClient, err = sftp.NewClient(sshClient); err == nil {
			// defer sftpClient.Close()

			srcFile, err := os.Open(sourceFile)
			if err != nil {
				log.Fatal(err)
			} else {
				defer srcFile.Close()

				fmt.Println("正在上传文件....")
				pathRemote := path.Join(conf.path, filepath.Base(sourceFile))

				dstFile, errUpload := sftpClient.Create(pathRemote)
				if nil != errUpload {
					log.Fatal(err)
				} else {
					defer dstFile.Close()

					ff, err := ioutil.ReadAll(srcFile)
					if err != nil {
						log.Fatal(err)
					}
					dstFile.Write(ff)
					fmt.Printf("[%s]上传到[%s@%s%s]\n", sourceFile, conf.username, conf.host, pathRemote)
				}
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}

type confLocal struct {
	DirLocal   string
	Host       string
	Port       int
	User       string
	Pwd        string
	DirRemote  string
	VersionGig int
}

type resultContent struct {
	Files        map[string]string
	VersionGig   int
	VersionSmall int
}

func deal(confDir string) {
	var conf confLocal
	confContent, _ := ioutil.ReadFile(confDir)
	err := json.Unmarshal(confContent, &conf)
	if err != nil {
		log.Fatal(err)
	}
	md5 := md5.New()

	dirPath := conf.DirLocal
	md5.Write([]byte(dirPath))
	MD5Str := hex.EncodeToString(md5.Sum(nil))
	dirPathTmp := filepath.Join(filepath.Dir(dirPath), fmt.Sprintf(".%s_%s", filepath.Base(dirPath), MD5Str))
	dirPathTmpFiles := filepath.Join(dirPathTmp, "tmp")

	os.MkdirAll(dirPathTmp, 0777)
	os.Mkdir(dirPathTmpFiles, 0777)

	pathResultFile := filepath.Join(dirPathTmp, "result.json")

	resultPrev := resultContent{}
	if isFileExists(pathResultFile) {
		jsonContent, _ := ioutil.ReadFile(pathResultFile)
		json.Unmarshal(jsonContent, &resultPrev)
	}
	files, _ := walkDir(dirPath)

	filesNew := make([]map[string]string, 0, 30)
	resultFiles := make(map[string]string)
	for _, file := range files {
		md5OfFile, _ := getFileMd5(file)
		if md5Old, ok := resultPrev.Files[file]; !ok || md5Old != md5OfFile {
			filePathRel, _ := filepath.Rel(dirPath, file)
			fileNew := filepath.Join(dirPathTmpFiles, filePathRel)
			os.MkdirAll(filepath.Dir(fileNew), 0777)
			copy(file, fileNew)
			filesNew = append(filesNew, map[string]string{
				"path": fileNew,
				"name": filePathRel,
			})
			resultFiles[file] = md5OfFile
			fmt.Println("修改文件[", fileNew, "]")
		}
	}
	resultPrev.Files = resultFiles

	if len(filesNew) > 0 {
		if resultPrev.VersionGig != conf.VersionGig {
			resultPrev.VersionGig = conf.VersionGig
			resultPrev.VersionSmall = 0
		} else {
			resultPrev.VersionSmall++
		}
		resultStr, _ := json.Marshal(resultPrev)
		ioutil.WriteFile(pathResultFile, resultStr, 0777)
		fmt.Println("生成结果文件[", pathResultFile, "]")

		versionCurrent := fmt.Sprintf("%d.%d", resultPrev.VersionGig, resultPrev.VersionSmall)
		pathVersionCurrent := path.Join(dirPathTmp, versionCurrent)

		os.MkdirAll(pathVersionCurrent, 0777)
		copy(pathResultFile, path.Join(pathVersionCurrent, "result.json"))

		zipName := time.Now().Format("20060102150405.zip")
		zipPath := filepath.Join(pathVersionCurrent, zipName)

		createZip(zipPath, filesNew)

		fmt.Println("生成zip文件[", zipPath, "]")

		sshConfig := serverConfig{
			host:     conf.Host,
			port:     conf.Port,
			username: conf.User,
			passwd:   conf.Pwd,
			path:     conf.DirRemote,
		}
		upload(zipPath, sshConfig)

		// 删除生成的临时文件夹
		os.RemoveAll(dirPathTmpFiles)

		var jsonUpdateFile = make(map[string][]map[string]string)
		pathUpdateFile := path.Join(dirPathTmp, "version.txt")

		var dataUpdateFile = make([]map[string]string, 0, 0)
		if isFileExists(pathUpdateFile) {
			jsonContent, _ := ioutil.ReadFile(pathUpdateFile)
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
		fmt.Println("生成更新文件[", pathUpdateFile, "]")

		upload(pathUpdateFile, sshConfig)
	} else {
		fmt.Println("没有要处理的更新文件")
	}
}

// {
//     "DirLocal": "本地路径",
//     "Host": "183.60.204.134",
//     "Port": 22,
//     "User": "用户名",
//     "Pwd": "密码",
//     "DirRemote": "远程路径",
//     "VersionGig": 1
// }
func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir)
	if err != nil {
		log.Fatal(err)
	} else {
		confDir := path.Join(dir, "conf.json")
		// confDir := "/Users/tonny/source/go/src/test/packageAndUpload/conf1.json"
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
