package utils

import (
	"log"
	"os"

	"github.com/toolkits/file"
)

func TaskCache(taskId, taskType string) (string, error) {
	file.InsureDir(TaskCacheDir)
	fileP := TaskCacheDir + "/" + taskId + "-" + taskType
	f, err := os.Create(fileP)
	if err != nil {
		log.Println("cache task failed:", err)
		return "", err
	}
	defer f.Close()
	return fileP, nil
}
