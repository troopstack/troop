package main

import (
	"flag"
	"fmt"
	"os"

	taskCache "github.com/troopstack/troop/src/modules/general/cache/task"
	"github.com/troopstack/troop/src/modules/general/database"
	"github.com/troopstack/troop/src/modules/general/http"
	"github.com/troopstack/troop/src/modules/general/rmq"
	"github.com/troopstack/troop/src/modules/general/rpc"
	"github.com/troopstack/troop/src/modules/general/scout"
	"github.com/troopstack/troop/src/modules/general/utils"
)

func main() {
	cfg := flag.String("c", "config.ini", "configuration file")
	version := flag.Bool("v", false, "show version")
	reset := flag.Bool("r", false, "reset generate key")

	flag.Parse()

	if *version {
		fmt.Println(utils.VERSION)
		os.Exit(0)
	}
	utils.ParseConfig(*cfg)

	logLevel := "info"
	if utils.Config().Debug.Enabled {
		logLevel = "debug"
	}
	utils.InitLog(logLevel)

	utils.InitRootDir()
	utils.InitDir()

	utils.RandomGenerateAES()

	if *reset || !utils.IsFile(utils.GeneralPriFilename) || !utils.IsFile(utils.GeneralPubFilename) {
		utils.KeyGenerate()
	}

	database.InitMySQL()

	taskCache.ReadTasksFromLocal()

	go utils.SignalListen()

	go rpc.Start()
	go http.Start()

	rmq.InitAmqp()
	scout.SendHandshakeMessage()

	// plugin
	utils.PluginCh = make(chan int, 1)
	go utils.InitPlugins()
	go func() {
		select {
		case <-utils.PluginCh:
			scout.SendUpdatePluginMessage()
		}
	}()

	select {}
}
