package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/troopstack/troop/src/modules/scout/rmq"
	"github.com/troopstack/troop/src/modules/scout/rpc"
	"github.com/troopstack/troop/src/modules/scout/run"
	"github.com/troopstack/troop/src/modules/scout/utils"
)

func main() {
	cfg := flag.String("c", "config.ini", "configuration file")
	version := flag.Bool("v", false, "show version")
	reset := flag.Bool("r", false, "reset generate key and reset auth")

	flag.Parse()

	if *version {
		fmt.Println(utils.VERSION)
		os.Exit(0)
	}
	utils.InitLocalIp()

	utils.ParseConfig(*cfg)

	logLevel := "info"
	if utils.Config().Debug.Enabled {
		logLevel = "debug"
		log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	}
	utils.InitLog(logLevel)

	utils.InitRootDir()
	utils.InitDir()

	utils.RandomGenerateAES()

	if *reset || !utils.IsFile(utils.ScoutPubFilename) || !utils.IsFile(utils.ScoutPriFilename) {
		utils.KeyGenerate()
	}

	utils.HandshakeChan = make(chan string, 1)
	utils.MqConnChan = make(chan string, 1)
	go rmq.SetupRMQ(true)
	go utils.CallHandshake()

	go func() {
		select {
		case handshakeResult := <-utils.HandshakeChan:
			if handshakeResult != "denied" {
				go rmq.ReSetupRMQ()

				select {
				case <-utils.MqConnChan:
					go rmq.MQBindTagQueue()
				}

			} else {
				log.Print("handshake is rejected")
				os.Exit(0)
			}
		}
	}()

	utils.HandshakeAccept = make(chan int, 1)
	go func() {
		select {
		case <-utils.HandshakeAccept:
			run.DownloadPlugin(false, "")
			rpc.HandleCacheTask()
			//err := run.RunScoutManager()
			//utils.FailOnError(err, "run scout manager error")
		}
	}()

	select {}
}
