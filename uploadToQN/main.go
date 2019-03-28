package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/qiniu/api.v7/auth/qbox"

	"github.com/qiniu/api.v7/storage"
)

var (
	accessKey = "rCyQK2Z_LmPf6uBFXpao9aJhTg1CMC7K1ghjUNM4"
	secretKey = "CUj14Kx03hQA4QVJvpvPpiZSRhGcl6kyR75NY8sX"
	bucket    = "xy3_bak"
	recordDir = "/tmp/upload2QN/"
)

func md5Hex(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// ProgressRecord data
type ProgressRecord struct {
	Progresses []storage.BlkputRet `json:"progresses"`
}

func isFileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
func main() {
	timeStart := time.Now()
	args := os.Args
	localFile := ""
	if len(args) > 1 {
		localFile = args[1]
	}
	if !isFileExists(localFile) {
		fmt.Printf("[%s]不存在", localFile)
		os.Exit(0)
	}
	key := filepath.Base(localFile)
	fmt.Printf("准备处理 %s\n", localFile)
	// localFile := "/Users/tonny/source/doc/xy3/bak/dump.6381.rdb.2019-02-16-16-38-52.rdb"
	// key := "dump.6381.rdb.2019-02-16-16-38-52.rdb"
	// 指定的进度文件保存目录，实际情况下，请确保该目录存在，而且只用于记录进度文件
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = true
	// 必须仔细选择一个能标志上传唯一性的 recordKey 用来记录上传进度
	// 我们这里采用 md5(bucket+key+local_path+local_file_last_modified)+".progress" 作为记录上传进度的文件名
	fileInfo, statErr := os.Stat(localFile)
	if statErr != nil {
		fmt.Println(statErr)
		return
	}
	fileSize := fileInfo.Size()
	fileLmd := fileInfo.ModTime().UnixNano()
	numBlock := storage.BlockCount(fileSize)
	recordKey := md5Hex(fmt.Sprintf("%s:%s:%s:%d", bucket, key, localFile, fileLmd)) + ".progress"
	mErr := os.MkdirAll(recordDir, 0755)
	if mErr != nil {
		fmt.Println("mkdir for record dir error,", mErr)
		return
	}
	recordPath := filepath.Join(recordDir, recordKey)
	progressRecord := ProgressRecord{}
	// 尝试从旧的进度文件中读取进度
	recordFp, openErr := os.Open(recordPath)
	if openErr == nil {
		progressBytes, readErr := ioutil.ReadAll(recordFp)
		if readErr == nil {
			mErr := json.Unmarshal(progressBytes, &progressRecord)
			if mErr == nil {
				// 检查context 是否过期，避免701错误
				for _, item := range progressRecord.Progresses {
					if storage.IsContextExpired(item) {
						fmt.Println(item.ExpiredAt)
						progressRecord.Progresses = make([]storage.BlkputRet, numBlock)
						break
					}
				}
			}
		}
		recordFp.Close()
	}
	if len(progressRecord.Progresses) == 0 {
		progressRecord.Progresses = make([]storage.BlkputRet, numBlock)
	}
	resumeUploader := storage.NewResumeUploader(&cfg)
	ret := storage.PutRet{}
	progressLock := sync.RWMutex{}
	lenUploaded := 0
	putExtra := storage.RputExtra{
		Progresses: progressRecord.Progresses,
		Notify: func(blkIdx int, blkSize int, ret *storage.BlkputRet) {
			progressLock.Lock()
			progressLock.Unlock()
			//将进度序列化，然后写入文件
			progressRecord.Progresses[blkIdx] = *ret
			lenUploaded++
			progressBytes, _ := json.Marshal(progressRecord)
			// fmt.Println("write progress file", numBlock, blkIdx, recordPath)
			// fmt.Printf("uploaded %d%\r", len(progressRecord.Progresses)*100/numBlock)
			fmt.Fprintf(os.Stdout, "正在上传文件 %d/%d, %d%%\r", lenUploaded, numBlock, lenUploaded*100/numBlock)
			wErr := ioutil.WriteFile(recordPath, progressBytes, 0644)
			if wErr != nil {
				fmt.Println("write progress file error,", wErr)
			}
		},
	}
	err := resumeUploader.PutFile(context.Background(), &ret, upToken, key, localFile, &putExtra)
	if err != nil {
		fmt.Println(err)
		return
	}
	//上传成功之后，一定记得删除这个进度文件
	os.Remove(recordPath)
	fmt.Println(ret.Key, "上传完成，总用时：", time.Since(timeStart))

	// 开始处理之前的老数据
	bucketManager := storage.NewBucketManager(mac, &cfg)
	marker := ""
	timeLastKeep := time.Now().Unix() - 60*60*6

	fmt.Printf("开始处理[ %v ]之前的数据\n", storage.ParsePutTime(timeLastKeep*10000*1000))
	// timeLastKeep := int64(1553755625)
	// fmt.Println(time.Now().Unix(), time.Now().Nanosecond(), timeLastKeep)
	// fmt.Println(storage.ParsePutTime(timeLastKeep * 10000 * 1000))
	// fmt.Println("")
	for {
		entries, _, nextMarker, hashNext, err := bucketManager.ListFiles(bucket, "", "", marker, 100)
		if err != nil {
			fmt.Println("list error,", err)
			break
		}
		//print entries
		for _, entry := range entries {
			timeFile := entry.PutTime / 10000 / 1000
			fmt.Println(entry.Key, storage.ParsePutTime(entry.PutTime))
			// fmt.Println(timeFile, timeFile < timeLastKeep)
			if timeFile < timeLastKeep {
				err := bucketManager.Delete(bucket, entry.Key)
				if err != nil {
					fmt.Println(err)
					return
				}

				fmt.Printf("删除文件[%s]\n", entry.Key)
			} else {
				fmt.Printf("文件[%s]不用处理\n", entry.Key)
			}
		}
		if hashNext {
			marker = nextMarker
		} else {
			//list end
			break
		}
	}
}
