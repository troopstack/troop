package model

import "fmt"

type ScoutHandRequest struct {
	Hostname     string
	IP           string
	ScoutVersion string
	Type         string
	PubKey       string
	AES          string
	OS           string
	Tags         []string
	Plugins      []string
}

type ScoutInfo struct {
	Hostname     string
	IP           string
	ScoutVersion string
	Type         string
	PubKey       string
	Status       string
	AES          string
	OS           string
	Tags         []string
	Plugins      []string
}

type ScoutTag struct {
	Name string
}

type ScoutUpdateInfo struct {
	LastUpdate  int64
	HandRequest *ScoutInfo
}

type ScoutHandResponse struct {
	Data           []byte
	Status         string
	Plugins        map[string]interface{}
	IgnoreCommands []string
}

type MatchedScout struct {
	Target     string
	TargetType string
	Tag        string
	OS         string
}

func (S *ScoutInfo) String() string {
	return fmt.Sprintf(
		"Hostname:%s, IP:%s, ScoutVersion:%s, Type:%s",
		S.Hostname,
		S.IP,
		S.ScoutVersion,
		S.Type,
	)
}
