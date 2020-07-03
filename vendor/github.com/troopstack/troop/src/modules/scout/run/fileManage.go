package run

import (
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"time"
)

type fileInfo struct {
	Name         string
	Size         int64
	Type         string
	LastModified time.Time
}

func FileList(prefix string) ([]*fileInfo, error) {
	files := []*fileInfo{}
	if runtime.GOOS == "windows" && prefix == "" {
		return GetDrives(), nil
	}
	fileInfoList, err := ioutil.ReadDir(prefix)
	if err != nil {
		log.Print(err)
		return files, err
	}
	for i := range fileInfoList {
		fileInfo := &fileInfo{
			Name:         fileInfoList[i].Name(),
			Size:         fileInfoList[i].Size(),
			LastModified: fileInfoList[i].ModTime(),
		}
		if fileInfoList[i].IsDir() {
			fileInfo.Type = `dir`
		} else {
			fileInfo.Type = `file`
		}
		files = append(files, fileInfo)
		//files[i] = fileInfo
	}
	return files, nil
}

func GetDrives() (r []*fileInfo) {
	i := 0
	for _, drive := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		f, err := os.Open(string(drive) + ":\\")
		if err == nil {
			fileInfo := &fileInfo{
				Name: f.Name(),
				Type: `dir`,
			}
			r = append(r, fileInfo)
			i += 1
			f.Close()
		}
	}
	return
}
