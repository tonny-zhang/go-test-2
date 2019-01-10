package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/howeyc/gopass"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type serverConfig struct {
	host              string
	port              int
	username          string
	path              string
	command           string
	msgCommandSuccess string
	msgCommandFail    string
	ext               string
}

func connect(user, password, host string, port int) (*sftp.Client, *ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, sshClient, err
	}

	return sftpClient, sshClient, nil
}
func errPrint(msg string) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", msg)
}

/*在可执行文件当前目录下新建data目录，直接双击可执行文件即可*/
func main() {
	conf := serverConfig{
		host:              "192.168.1.108",
		port:              22,
		username:          "tonny",
		path:              "/home/tonny/test/",
		command:           "/home/tonny/test.sh",
		msgCommandSuccess: "重启成功",
		msgCommandFail:    "重启失败",
		ext:               ".json",
	}

	// var username, pwd string

	// fmt.Print("请输入用户名: ")
	// fmt.Scan(&username)
	fmt.Print("请输入密码: ")
	pass, errPwd := gopass.GetPasswdMasked()
	if errPwd != nil {
		errPrint("读取密码错误")

		os.Exit(0)
	}
	var (
		err        error
		sftpClient *sftp.Client
		sshClient  *ssh.Client
	)

	sftpClient, sshClient, err = connect(conf.username, string(pass), conf.host, conf.port)
	if err != nil {
		fmt.Println("密码错误！")
		os.Exit(0)
	} else {
		defer sftpClient.Close()
	}

	var dirCurrent, _ = os.Getwd()
	datadir := path.Join(dirCurrent, "data")

	if info, err := os.Stat(datadir); !os.IsNotExist(err) && info.IsDir() {
		ext := conf.ext
		files, err := ioutil.ReadDir(datadir)
		if err == nil {
			for _, file := range files {
				filename := file.Name()
				pathLocal := path.Join(datadir, filename)
				if info1, err1 := os.Stat(pathLocal); nil == err1 && !info1.IsDir() {
					if ext != "" && path.Ext(filename) != ext {
						continue
					}
					srcFile, err := os.Open(pathLocal)
					if err != nil {
						log.Fatal(err)
					} else {
						defer srcFile.Close()

						pathRemote := path.Join(conf.path, filename)

						dstFile, errUpload := sftpClient.Create(pathRemote)
						if nil != errUpload {
							log.Fatal(err)
						} else {
							defer dstFile.Close()

							buf := make([]byte, 1024)
							for {
								n, _ := srcFile.Read(buf)
								if n == 0 {
									break
								}
								dstFile.Write(buf)
							}
							fmt.Printf("%s\n", filename)
						}
					}
				} else {
					fmt.Println("暂时不支持子目录")
				}
			}

			command := conf.command
			if command != "" {
				if session, err := sshClient.NewSession(); err != nil {
					errPrint("错误")
				} else {
					defer session.Close()
					// session.Stdout = os.Stdout
					// session.Stderr = os.Stderr
					if nil != session.Run(command) {
						errPrint(conf.msgCommandFail)
					} else {
						fmt.Println("\n" + conf.msgCommandSuccess)
					}
				}
			}
		} else {
			errPrint("读取data目录出现错误")
		}
	} else {
		errPrint("当前目录下没有data目录")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n\n回车退出...")
	reader.ReadByte()
	os.Exit(0)
}
