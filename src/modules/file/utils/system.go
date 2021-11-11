package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path"
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
func IsDir(dir string) bool {
	if info, err := os.Stat(dir); err == nil {
		return info.IsDir()
	}
	return false
}

// 如果目录不存在则创建
func CreateDir(dir string) error {
	if !IsDir(dir) {
		err := os.Mkdir(dir, 0766)
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

// 解压 tar.gz
func DeCompress(srcFilePath string, destDirPath string) error {
	os.Mkdir(destDirPath, os.ModePerm)

	fr, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer fr.Close()

	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if hdr.Typeflag != tar.TypeDir {
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)

			fw, _ := os.Create(destDirPath + "/" + hdr.Name)
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, tr)
			if err != nil {
				return err
			}
			fw.Close()
		}
	}
	return nil
}
