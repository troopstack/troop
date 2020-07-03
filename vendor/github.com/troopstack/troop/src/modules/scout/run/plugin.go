package run

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/troopstack/troop/src/model"
	"github.com/troopstack/troop/src/modules/scout/rpc"
	"github.com/troopstack/troop/src/modules/scout/utils"

	"github.com/mitchellh/mapstructure"
	"github.com/toolkits/file"
)

type pluginS struct {
	Windows map[string]interface{} `json:"windows"`
	Linux   map[string]interface{} `json:"linux"`
}

type pluginInfo struct {
	Url string `json:"url"`
}

func PluginAction(scoutPluginRequest model.ScoutPluginRequest) (string, error) {
	var result = ""
	var err error
	switch scoutPluginRequest.Action {
	case "update_plugins":
		utils.Plugins = scoutPluginRequest.Plugins
		result = DownloadPlugin(scoutPluginRequest.Args == "cover", scoutPluginRequest.Plugin)
	case "config.update":
		var configPath string
		configPath, err = DownloadConfigFile(scoutPluginRequest.TaskId, scoutPluginRequest.FileUrl)
		if scoutPluginRequest.Plugin == "scout_manager" {
			var cacheFile string
			cacheFile, err = utils.TaskCache(scoutPluginRequest.TaskId, "update")
			if err == nil {
				result, err = execPluginAction(scoutPluginRequest.Plugin, scoutPluginRequest.Action,
					configPath)
				if err != nil {
					defer os.Remove(cacheFile)
				}
			}
		} else {
			result, err = execPluginAction(scoutPluginRequest.Plugin, scoutPluginRequest.Action,
				configPath)
		}
	case "update":
		if scoutPluginRequest.Plugin == "scout_manager" {
			var cacheFile string
			cacheFile, err = utils.TaskCache(scoutPluginRequest.TaskId, "update")
			if err == nil {
				result, err = execPluginAction(scoutPluginRequest.Plugin, scoutPluginRequest.Action,
					scoutPluginRequest.Args, utils.Root)
				if err != nil {
					defer os.Remove(cacheFile)
				}
			}
		}
	default:
		result, err = execPluginAction(scoutPluginRequest.Plugin, scoutPluginRequest.Action,
			scoutPluginRequest.Args)
	}
	return result, err
}

func formatVersionNum(version string) string {
	version = strings.TrimSpace(version)
	if strings.HasPrefix(version, "v") {
	} else if strings.HasPrefix(version, "V") {
		version = strings.Replace(version, "V", "v", 1)
	} else {
		version = "v" + version
	}
	return version
}

func pluginFilePath(pluginName string) (string, error) {
	filepath := path.Join(utils.Root, utils.PluginParentDir, pluginName, pluginName)
	if runtime.GOOS == "windows" {
		filepath = filepath + ".exe"
	}
	exists := utils.IsFile(filepath)
	if !exists {
		return filepath, errors.New("plugin not exists")
	}
	return filepath, nil
}

func execPluginAction(pluginName string, args ...string) (string, error) {
	pluginPath, err := pluginFilePath(pluginName)
	if err == nil {
		result, runNormal := OrderRunStart(pluginPath, nil, "", args...)
		if !runNormal {
			log.Printf("plugin exec action %s failed : %s", args, result.Error)
		}
		if result.Error != "" {
			err = errors.New(result.Error)
		}
		return result.Stdout, err
	}
	return "", err
}

func pluginWriteConfigIni(plugin string) {
	if plugin == "" {
		return
	}
	configPlugins := utils.GetString("plugin", "plugins")
	newConfigPluginsSplit := []string{}
	if configPlugins != "" {
		configPlugins = strings.TrimSpace(configPlugins)
		configPluginsSplit := strings.Split(configPlugins, ",")
		newConfigPluginsSplit = append(newConfigPluginsSplit, configPluginsSplit...)
	}

	newConfigPluginsSplit = append(newConfigPluginsSplit, plugin)

	newConfigPlugins := strings.Join(newConfigPluginsSplit, ",")
	ok := utils.WriteIni("plugin", "plugins", newConfigPlugins)
	if ok {
		hostname, errH := utils.Hostname()
		if errH == nil {
			data := model.UpdateScoutHavePluginRequest{
				Hostname: hostname,
				Plugins:  newConfigPluginsSplit,
			}
			rpc.SendScoutHavePlugins(data)
		}
	}
}

func DownloadPlugin(cover bool, plugin string) string {
	usePlugins := utils.Config().Plugin.Plugins
	plugins := utils.Plugins
	pluginUpdateResults := []string{}
	over := false
	for i := range usePlugins {
		if over {
			break
		}
		pluginNameAndVersion := strings.Split(usePlugins[i], ":")
		pluginName := pluginNameAndVersion[0]
		writeConfigPlugin := ""
		// 判断是否指定插件名
		if plugin != "" {
			specifyPluginNameAndVersion := strings.Split(plugin, ":")
			specifyPluginName := specifyPluginNameAndVersion[0]
			if specifyPluginName != pluginName {
				if i < len(usePlugins)-1 {
					continue
				} else {
					writeConfigPlugin = plugin
					pluginNameAndVersion = specifyPluginNameAndVersion
					pluginName = specifyPluginName
				}
			} else {
				over = true
				pluginNameAndVersion = specifyPluginNameAndVersion
			}
		}
		pluginVersion := ""
		// 默认拉取最新版本的插件，如指定版本则拉取指定版本
		if len(pluginNameAndVersion) > 1 {
			pluginVersion = formatVersionNum(pluginNameAndVersion[1])
		}
		if plugin, ok := plugins[pluginName]; ok {
			pluginS := pluginS{}
			err := mapstructure.Decode(plugin, &pluginS)
			if err != nil {
				log.Printf("plugin: %s abnormal", pluginName)
				continue
			}
			var pluginVersions = make(map[string]interface{})
			if runtime.GOOS == "windows" {
				pluginVersions = pluginS.Windows
			} else {
				pluginVersions = pluginS.Linux
			}
			if len(pluginVersions) == 0 {
				errLog := fmt.Sprintf("%s plugin: %s does not exist", runtime.GOOS, pluginName)
				pluginUpdateResults = append(pluginUpdateResults, errLog)
				continue
			}
			if pluginVersion == "" {
				versions := []string{}
				for version := range pluginVersions {
					versions = append(versions, version)
				}
				versions = utils.VersionCompare(versions)
				if len(versions) > 0 {
					pluginVersion = versions[0]
				} else {
					errLog := fmt.Sprintf("%s plugin: %s does not exist", runtime.GOOS, pluginName)
					pluginUpdateResults = append(pluginUpdateResults, errLog)
					continue
				}
			}
			currentVersion, err := execPluginAction(pluginName, "version")
			if !cover {
				if err == nil {
					currentVersion = formatVersionNum(currentVersion)
					if currentVersion == formatVersionNum(pluginVersion) {
						pluginUpdateResults = append(pluginUpdateResults,
							fmt.Sprintln(pluginName+":", currentVersion, "already exists"))
						continue
					}
				}
			}
			pluginInfoDir := pluginVersions[pluginVersion]
			pluginInfo := pluginInfo{}
			err = mapstructure.Decode(pluginInfoDir, &pluginInfo)
			if err != nil || pluginInfo.Url == "" {
				log.Printf("plugin: %s, version: %s not exists", pluginName, pluginVersion)
				errLog := fmt.Sprintf("version: %s does not exist in the repository", pluginVersion)
				pluginUpdateResults = append(pluginUpdateResults, fmt.Sprintln(pluginName, errLog))
				continue
			}
			log.Printf("start download plugin: %s, version: %s", pluginName, pluginVersion)
			err = download(pluginName, pluginInfo.Url)
			if err != nil {
				errLog := fmt.Sprintf("version: %s download failed: %s", pluginVersion, err.Error())
				pluginUpdateResults = append(pluginUpdateResults, fmt.Sprintln(pluginName, errLog))
			} else {
				successLog := fmt.Sprintf("version: %s -> %s",
					formatVersionNum(currentVersion), formatVersionNum(pluginVersion))
				pluginUpdateResults = append(pluginUpdateResults,
					fmt.Sprintln(pluginName, successLog))
				if writeConfigPlugin != "" && pluginName != "scout_manager" {
					go pluginWriteConfigIni(writeConfigPlugin)
				}
			}
		} else {
			log.Printf("plugin: %s not exists", usePlugins[i])
		}
	}
	return strings.Join(pluginUpdateResults, "")
}

func download(pluginName, url string) error {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Http-Token", utils.Config().General.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		utils.FailOnError(err, "plugin download failed")
		return err
	}

	if res.StatusCode == 401 {
		log.Print("plugin download failed: invalid token.")
		return err
	}

	pluginDir := path.Join(utils.PluginParentDir, pluginName)

	err = file.InsureDir(pluginDir)
	if err != nil {
		utils.FailOnError(err, "insure dir failed")
		return err
	}

	fileP := path.Join(pluginDir, file.Basename(url))

	f, err := os.OpenFile(fileP, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		utils.FailOnError(err, "create file failed")
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		utils.FailOnError(err, "io copy failed")
		return err
	}

	return nil
}

func DownloadConfigFile(taskId, url string) (string, error) {
	saveDir := path.Join(utils.FileCacheDir, taskId)
	err := file.InsureDir(saveDir)

	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Http-Token", utils.Config().General.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Print(err.Error())
		return "", err
	}

	if res.StatusCode == 401 {
		return "", errors.New("file download failed: invalid token")
	}

	fileP := path.Join(saveDir, file.Basename(url))

	f, err := os.Create(fileP)
	defer f.Close()
	if err != nil {
		log.Print(err.Error())
		return "", err
	}
	io.Copy(f, res.Body)
	return fileP, nil
}
