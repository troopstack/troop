package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/troopstack/troop/src/model"
)

func SaveGeneralPki(PubKey string) error {
	if !isDir(GeneralPkiRoot) {
		err := os.Mkdir(GeneralPkiRoot, 0666)
		FailOnError(err, "")
		return err
	}
	PubKeyBytes := []byte(PubKey)
	err := ioutil.WriteFile(GeneralPubFilename, PubKeyBytes, 0666)
	FailOnError(err, "")
	return err
}

func ReadScoutPubKey() (string, bool) {
	if IsFile(ScoutPubFilename) {
		ScoutPub, err := ioutil.ReadFile(ScoutPubFilename)
		if err != nil {
			log.Fatalf("Scout Public read failed：%s", err)
			return "", false
		}
		return string(ScoutPub), true
	} else {
		return "", false
	}
}

func ReadScoutPriKey() (string, bool) {
	if IsFile(ScoutPriFilename) {
		ScoutPub, err := ioutil.ReadFile(ScoutPriFilename)
		if err != nil {
			log.Fatalf("Scout Private read failed：%s", err)
			return "", false
		}
		return string(ScoutPub), true
	} else {
		return "", false
	}
}

func SaveScoutPubKey(PubKey string) error {
	if !isDir(ScoutPkiRoot) {
		err := os.Mkdir(ScoutPkiRoot, 0666)
		FailOnError(err, "")
		if err != nil {
			return err
		}
	}
	PubKeyBytes := []byte(PubKey)
	err := ioutil.WriteFile(ScoutPubFilename, PubKeyBytes, 0666)
	FailOnError(err, "")
	return err
}

func SaveScoutPriKey(PriKey string) error {
	if !isDir(ScoutPkiRoot) {
		err := os.Mkdir(ScoutPkiRoot, 0666)
		FailOnError(err, "")
		if err != nil {
			return err
		}
	}
	PriKeyBytes := []byte(PriKey)
	err := ioutil.WriteFile(ScoutPriFilename, PriKeyBytes, 0666)
	FailOnError(err, "")
	return err
}

func SaveGeneralInfo(data []byte, ScoutPriKey string) error {
	// 私钥解密
	res, err := RsaDecrypt(data, []byte(ScoutPriKey))
	if err != nil {
		return err
	}
	UnCipherText := UnCipherText{}
	err = json.Unmarshal(res, &UnCipherText)
	if err != nil {
		return err
	}
	GeneralAES = UnCipherText.AES
	err = SaveGeneralPki(UnCipherText.GeneralPub)
	if err != nil {
		return err
	}

	go func() {
		var resp model.RpcInitDataResponse
		CallGeneral("Scout.GeneralInitData", nil, &resp)
		Plugins = resp.Plugins
		GeneralIgnoreCommands = resp.IgnoreCommands
		HandshakeAccept <- 1
	}()

	return nil
}
