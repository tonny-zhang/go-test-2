package uploader

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"test/packageAndUpload/config"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type serverConfig struct {
	host     string
	port     int
	username string
	passwd   string
	path     string
}

func printErrAndWaitExit(err interface{}) {
	log.Fatal(err)
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n\n回车退出...")
	reader.ReadByte()
	os.Exit(0)
}
func uploadSSH(sourceFile string, conflocal config.ConfLocal) {
	conf := serverConfig{
		host:     conflocal.Host,
		port:     conflocal.Port,
		username: conflocal.User,
		passwd:   conflocal.Pwd,
		path:     conflocal.DirRemote,
	}
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
				printErrAndWaitExit(err)
			} else {
				defer srcFile.Close()

				// fmt.Println("正在上传文件....")
				pathRemote := path.Join(conf.path, filepath.Base(sourceFile))

				dstFile, errUpload := sftpClient.Create(pathRemote)
				if nil != errUpload {
					printErrAndWaitExit(err)
				} else {
					defer dstFile.Close()

					ff, err := ioutil.ReadAll(srcFile)
					if err != nil {
						printErrAndWaitExit(err)
					}

					// 使用进度条上传
					countTotal := len(ff)
					lenCache := 1024 * 543

					tUpload := time.Now()
					fmt.Fprintf(os.Stdout, "正在上传文件 %d/%d, %d%%\r", 0, countTotal, 0)
					for indexStart := 0; indexStart < countTotal; {
						indexEnd := indexStart + lenCache
						if indexEnd > countTotal {
							indexEnd = countTotal
						}
						dstFile.Write(ff[indexStart:indexEnd])
						perc := indexEnd * 100 / countTotal
						fmt.Fprintf(os.Stdout, "正在上传文件 %d/%d, %d%%\r", indexEnd, countTotal, perc)
						indexStart = indexEnd
					}

					// 直接上传
					// dstFile.Write(ff)
					fmt.Printf("[%s]上传到[%s@%s%s], 用时%v\n", sourceFile, conf.username, conf.host, pathRemote, time.Since(tUpload))
				}
			}
		} else {
			printErrAndWaitExit(err)
		}
	} else {
		printErrAndWaitExit(err)
	}
}
