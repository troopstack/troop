package utils

import (
	"os"
)

// 判断文件或者文件夹是否存在，一般判断第一个参数即可，第二个参数可以忽略，或者严谨一些，把err日志记录起来
func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err == nil {
		return true, nil //文件或者文件夹存在
	}
	if os.IsNotExist(err) {
		return false, nil //不存在
	}
	return false, err //不存在，这里的err可以查到具体的错误信息
}

// 判断目录是否存在
func isDir(dir string) bool {
	if info, err := os.Stat(dir); err == nil {
		return info.IsDir()
	}
	return false
}

// 如果目录不存在则创建
func CreateDir(dir string) error {
	if !isDir(dir) {
		err := os.Mkdir(dir, 0666)
		FailOnError(err, "")
		return err
	}
	return nil
}

// 判断文件是否存在
func IsFile(file string) bool {
	existed := true
	if _, err := os.Stat(file); os.IsNotExist(err) {
		existed = false
	}
	return existed
}
