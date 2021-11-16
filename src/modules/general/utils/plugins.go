package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/toolkits/file"
)

type pluginUploadResponse struct {
	Url string
}

var (
	Plugins         = make(map[string]interface{})
	PluginTarName   = "plugins.tar.gz"
	PluginCh        chan int
	pluginParentDir string
	pluginDirFile   string
	pluginFMUrl     string
)

func download() error {
	// 从Git拉取插件
	if file.IsExist(pluginDirFile) {
		// git pull
		cmd := exec.Command("git", "pull", "--ff-only")
		cmd.Dir = pluginDirFile
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("git pull in dir:%s fail. error: %s", pluginDirFile, err)
			return err
		}
	} else {
		// git clone
		cmd := exec.Command("git", "clone", Config().Plugin.Git, file.Basename(pluginDirFile))
		cmd.Dir = pluginParentDir
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("git clone in dir:%s fail. error: %s", pluginParentDir, err)
			return err
		}
	}
	return nil
}

func sendToFileSystem(pluginsDirName string) error {
	// 推送插件到File服务
	fileServerUrl := Config().File.Address + "/plugin/upload"

	existed := isDir(pluginDirFile)
	if !existed {
		fmt.Println("error: dir", pluginDirFile, "not exists")
		return errors.New("plugin not exists")
	}
	pluginTarFile := path.Join(pluginParentDir, PluginTarName)
	err := Compress(pluginDirFile, pluginTarFile, false)
	if err != nil {
		log.Printf("compress plugins dir failed: %s", err)
		return err
	}

	fileByte, err := ioutil.ReadFile(pluginTarFile)

	FailOnError(err, "plugin read failed")

	data := make(map[string]interface{})
	data["file"] = fileByte
	data["file_name"] = PluginTarName
	data["plugins_pathname"] = pluginsDirName

	bytesData, _ := json.Marshal(data)

	payload := bytes.NewReader(bytesData)

	req, _ := http.NewRequest("POST", fileServerUrl, payload)

	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Add("Http-Token", Config().Http.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Print("Error: General can not connection file server.")
		return err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode == 401 {
		fmt.Printf("General connection file server failed. %s", string(body))
		return err
	}
	pluginUploadResponse := pluginUploadResponse{}

	err = json.Unmarshal(body, &pluginUploadResponse)
	if err != nil {
		log.Println(err.Error())
		log.Println(string(body))
	} else {
		pluginFMUrl = pluginUploadResponse.Url
	}
	return nil
}

func InitPlugins() {
	// 初始化插件
	if Config().Plugin.Enabled {
		dir := "plugins"
		parentDir := path.Join(Root, dir)
		file.InsureDir(parentDir)
		pluginParentDir = parentDir

		if Config().Plugin.Git != "" {
			dirName := strings.Split(file.Basename(Config().Plugin.Git), ".git")[0]
			pluginDirFile = parentDir + "/" + dirName

			err := download()
			if err != nil {
				log.Print("download plugins from git failed")
				log.Print(err)
				go pluginChSend(0)
				return
			}

			err = sendToFileSystem(dirName)

			if err != nil {
				log.Print("send plugins to file system failed")
				log.Print(err)
				go pluginChSend(0)
				return
			}

			plugins := make(map[string]interface{})

			pluginDirs, err := ioutil.ReadDir(pluginDirFile)
			if err != nil {
				FailOnError(err, "plugin dir failed")
				return
			}
			for _, pluDir := range pluginDirs {
				if pluDir.Name() == ".git" {
					continue
				}
				pluginInfo := make(map[string]map[string]interface{})
				pluginInfo["windows"] = make(map[string]interface{})
				pluginInfo["linux"] = make(map[string]interface{})
				if pluDir.IsDir() {
					vDir := path.Join(pluginDirFile, pluDir.Name())
					if isDir(vDir) {
						pluginFiles, readPluginDirErr := ioutil.ReadDir(vDir)
						if readPluginDirErr != nil {
							FailOnError(readPluginDirErr, vDir+" dir failed")
							continue
						}

						for _, d := range pluginFiles {
							if IsFile(path.Join(vDir, d.Name())) {
								// 文件名
								fileName := d.Name()
								// 文件名结尾
								fileSuffix := path.Ext(fileName)
								// 文件名前缀
								fileNameOnly := fileName
								// 判断文件名结尾是否为数字，如果为数字，则为版本号，后缀为空
								_, nErr := strconv.Atoi(strings.Replace(fileSuffix, ".", "", 1))
								if nErr != nil {
									fileNameOnly = strings.TrimSuffix(fileNameOnly, fileSuffix)
								}
								// 用`-`分割文件名，文件名格式为`[pluginName]-[os]-[version].[suffix]`，suffix为可选
								fnMap := strings.Split(fileNameOnly, "-")
								if len(fnMap) >= 3 {
									pOs := fnMap[1]
									if pOs == "win" {
										pOs = "windows"
									}
									version := fnMap[2]
									pluginInfoMap := make(map[string]interface{})
									// 文件下载路径
									pluginInfoMap["url"] = Config().File.Address + path.Join("/file/download",
										pluginFMUrl, pluDir.Name(), fileName)
									pluginInfo[pOs][version] = pluginInfoMap

								} else {
									FailOnError(err, fileName+" file name format error")
									continue
								}
							}
						}
					}

					//winDir := path.Join(pluginDirFile, pluDir.Name(), "windows")
					//if isDir(winDir) {
					//	winVersions, err := ioutil.ReadDir(winDir)
					//	if err != nil {
					//		FailOnError(err, winDir+" dir failed")
					//		continue
					//	}
					//	pluginWinVersionInfo := make(map[string]interface{})
					//	for _, winVersionDir := range winVersions {
					//		pluExecFile := path.Join(winDir, winVersionDir.Name(), pluDir.Name()+".exe")
					//		if IsFile(pluExecFile) {
					//			pluginWinInfo := make(map[string]interface{})
					//			pluginWinInfo["url"] = Config().File.Address + path.Join("/file/download", pluginFMUrl,
					//				pluDir.Name(), "windows", winVersionDir.Name(), pluDir.Name()+".exe")
					//			pluginWinVersionInfo[winVersionDir.Name()] = pluginWinInfo
					//		}
					//	}
					//	pluginInfo["windows"] = pluginWinVersionInfo
					//}
					//linuxDir := path.Join(pluginDirFile, pluDir.Name(), "linux")
					//if isDir(linuxDir) {
					//	linuxVersions, err := ioutil.ReadDir(linuxDir)
					//	if err != nil {
					//		FailOnError(err, linuxDir+" dir failed")
					//		continue
					//	}
					//	pluginLinuxVersionInfo := make(map[string]interface{})
					//	for _, LinuxVersionDir := range linuxVersions {
					//		pluExecFile := path.Join(linuxDir, LinuxVersionDir.Name(), pluDir.Name())
					//		if IsFile(pluExecFile) {
					//			pluginLinuxInfo := make(map[string]interface{})
					//			pluginLinuxInfo["url"] = Config().File.Address + path.Join("/file/download", pluginFMUrl,
					//				pluDir.Name(), "linux", LinuxVersionDir.Name(), pluDir.Name())
					//			pluginLinuxVersionInfo[LinuxVersionDir.Name()] = pluginLinuxInfo
					//		}
					//	}
					//	pluginInfo["linux"] = pluginLinuxVersionInfo
					//}
					plugins[pluDir.Name()] = pluginInfo
				}
			}
			Plugins = plugins
			go pluginChSend(1)

		} else {
			go pluginChSend(0)
			log.Printf("plugins git path failed")
		}
	} else {
		go pluginChSend(0)
		log.Printf("plugin not enabled")
	}
}

func pluginChSend(data int) {
	PluginCh <- data
}
