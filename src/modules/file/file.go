package main

import (
	"flag"

	"github.com/troopstack/troop/src/modules/file/http"
	"github.com/troopstack/troop/src/modules/file/utils"
)

func main() {
	cfg := flag.String("c", "config.ini", "configuration file")
	flag.Parse()
	utils.ParseConfig(*cfg)

	logLevel := "info"
	if utils.Config().Debug.Enabled {
		logLevel = "debug"
	}
	utils.InitLog(logLevel)

	utils.InitRootDir()
	utils.InitDir()

	go http.Start()

	select {}
}
