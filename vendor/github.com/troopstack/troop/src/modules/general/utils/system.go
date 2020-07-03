package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

// 将文件或目录打包成 .tar 文件
// src 是要打包的文件或目录的路径
// dstTar 是要生成的 .tar 文件的路径
// failIfExist 标记如果 dstTar 文件存在，是否放弃打包，如果否，则会覆盖已存在的文件
func Compress(src string, dstTar string, failIfExist bool) (err error) {
	// 清理路径字符串
	src = path.Clean(src)

	// 判断要打包的文件或目录是否存在
	if exists, _ := FileExists(src); !exists {
		return errors.New("要打包的文件或目录不存在：" + src)
	}

	// 判断目标文件是否存在
	if exists, _ := FileExists(dstTar); exists {
		if failIfExist { // 不覆盖已存在的文件
			return errors.New("目标文件已经存在：" + dstTar)
		} else { // 覆盖已存在的文件
			if er := os.Remove(dstTar); er != nil {
				return er
			}
		}
	}

	// 创建空的目标文件
	fw, er := os.Create(dstTar)
	if er != nil {
		return er
	}
	defer fw.Close()

	gw := gzip.NewWriter(fw)
	defer func() {
		// 这里要判断 gw 是否关闭成功，如果关闭失败，则 gz 文件可能不完整
		if er := gw.Close(); er != nil {
			err = er
		}
	}()

	// 创建 tar.Writer，执行打包操作
	tw := tar.NewWriter(gw)
	defer func() {
		// 这里要判断 tw 是否关闭成功，如果关闭失败，则 tar.gz 文件可能不完整
		if er := tw.Close(); er != nil {
			err = er
		}
	}()

	// 获取文件或目录信息
	fi, er := os.Stat(src)
	if er != nil {
		return er
	}

	// 获取要打包的文件或目录的所在位置和名称
	// srcBase, srcRelative := path.Split(filepath.Clean(src))

	srcBase := filepath.Dir(filepath.Clean(src))
	srcRelative := filepath.Base(filepath.Clean(src))

	// 开始打包
	if fi.IsDir() {
		tarDir(srcBase, srcRelative, tw, fi)
	} else {
		tarFile(srcBase, srcRelative, tw, fi)
	}

	return nil
}

// 因为要执行遍历操作，所以要单独创建一个函数
func tarDir(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := path.Join(srcBase, srcRelative)

	// 获取 srcFull 下的文件或子目录列表
	fis, er := ioutil.ReadDir(srcFull)
	if er != nil {
		return er
	}

	// 开始遍历
	for _, fi := range fis {
		if fi.IsDir() {
			if fi.Name() == ".git" {
				continue
			}
			tarDir(srcBase, path.Join(srcRelative, fi.Name()), tw, fi)
		} else {
			tarFile(srcBase, path.Join(srcRelative, fi.Name()), tw, fi)
		}
	}

	// 写入目录信息
	if len(srcRelative) > 0 {
		hdr, er := tar.FileInfoHeader(fi, "")
		if er != nil {
			return er
		}
		hdr.Name = srcRelative

		hdr.Format = tar.FormatGNU

		if er = tw.WriteHeader(hdr); er != nil {
			return er
		}
	}

	return nil
}

// 因为要在 defer 中关闭文件，所以要单独创建一个函数
func tarFile(srcBase, srcRelative string, tw *tar.Writer, fi os.FileInfo) (err error) {
	// 获取完整路径
	srcFull := path.Join(srcBase, srcRelative)

	// 写入文件信息
	hdr, er := tar.FileInfoHeader(fi, "")
	if er != nil {
		return er
	}
	hdr.Name = srcRelative
	hdr.Format = tar.FormatGNU

	if er = tw.WriteHeader(hdr); er != nil {
		return er
	}

	// 打开要打包的文件，准备读取
	fr, er := os.Open(srcFull)
	if er != nil {
		return er
	}
	defer fr.Close()

	// 将文件数据写入 tw 中
	if _, er = io.Copy(tw, fr); er != nil {
		return er
	}
	return nil
}
