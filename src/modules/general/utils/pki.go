package utils

import (
	"io/ioutil"
	"log"
	"os"
)

func SaveScoutPki(Hostname string, PubKey string) error {
	if !isDir(ScoutPkiRoot) {
		err := os.Mkdir(ScoutPkiRoot, 0666)
		FailOnError(err, "")
		return err
	}
	filename := ScoutPkiRoot + "/" + Hostname
	PubKeyBytes := []byte(PubKey)
	err := ioutil.WriteFile(filename, PubKeyBytes, 0666)
	FailOnError(err, "")
	return err
}

func ReadScoutPubKey(Hostname string) (string, bool) {
	filename := ScoutPkiRoot + "/" + Hostname
	if IsFile(filename) {
		GeneralPub, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Scout Public read failed：%s", err)
			return "", false
		}
		return string(GeneralPub), true
	} else {
		return "", false
	}

}

func SaveGeneralPubKey(PubKey string) error {
	if !isDir(GeneralPkiRoot) {
		err := os.Mkdir(GeneralPkiRoot, 0666)
		FailOnError(err, "")
		if err != nil {
			return err
		}
	}
	PubKeyBytes := []byte(PubKey)
	err := ioutil.WriteFile(GeneralPubFilename, PubKeyBytes, 0666)
	FailOnError(err, "")
	return err
}

func SaveGeneralPriKey(PriKey string) error {
	if !isDir(GeneralPkiRoot) {
		err := os.Mkdir(GeneralPkiRoot, 0666)
		FailOnError(err, "")
		if err != nil {
			return err
		}
	}
	PriKeyBytes := []byte(PriKey)
	err := ioutil.WriteFile(GeneralPriFilename, PriKeyBytes, 0666)
	FailOnError(err, "")
	return err
}
