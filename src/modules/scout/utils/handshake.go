package utils

import (
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/troopstack/troop/src/model"
)

var HandshakeChan chan string
var HandshakeAccept chan int

func CallHandshake() bool {
	hostname, err := Hostname()
	if err != nil {
		hostname = fmt.Sprintf("error:%s", err.Error())
		panic(err)
	}
	ScoutPubKey, isHave := ReadScoutPubKey()
	if !isHave {
		panic(errors.New("scout Public read failed"))
	}

	ScoutPriKey, isHave := ReadScoutPriKey()
	if !isHave {
		panic(errors.New("scout Private read failed"))
	}

	req := model.ScoutHandRequest{
		Hostname:     hostname,
		IP:           IP(),
		ScoutVersion: VERSION,
		Type:         "server",
		PubKey:       ScoutPubKey,
		AES:          AES,
		OS:           runtime.GOOS,
		Tags:         Config().Host.Tag,
		Plugins:      Config().Plugin.CustomPlugins,
	}
	var resp model.ScoutHandResponse
	log.Print("Handshake with general")
	callOk := CallGeneral("Scout.Handshake", req, &resp)
	if !callOk {
		return false
	}

	if resp.Status == "accepted" {
		err := SaveGeneralInfo(resp.Data, ScoutPriKey)
		if err != nil {
			panic(err)
		}
		Plugins = resp.Plugins
		GeneralIgnoreCommands = resp.IgnoreCommands
	}
	log.Printf("Handshake successfully. status: %s.", resp.Status)

	go func() {
		HandshakeChan <- resp.Status
	}()

	return true
}
