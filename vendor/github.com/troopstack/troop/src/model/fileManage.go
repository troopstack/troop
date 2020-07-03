package model

import "time"

type FileInfo struct {
	Name         string
	Size         int64
	Type         string
	LastModified time.Time
}

type FileManageScoutInfo struct {
	TaskId    string
	Scout     string
	ScoutType string
	Result    string
	Error     string
	Status    string
}

type FileManageResultRequest struct {
	TaskId    string
	Scout     string
	ScoutType string
	Result    string
	Error     string
	Complete  bool
}

type FileManageRequest struct {
	Action     string `json:"action" form:"action"`
	Prefix     string `json:"prefix" form:"prefix"`
	Target     string `json:"target" form:"target" binding:"required"`
	TargetType string `json:"target_type" form:"target_type" default:"server"`
	Tag        string `json:"tag" form:"tag"`
	OS         string `json:"os" form:"os"`
	Detach     bool   `json:"detach" form:"detach" default:"false"`
	Timeout    int    `json:"timeout" form:"timeout"`
}

type FileManageTaskRequest struct {
	TaskId string
	Action string
	Prefix string
}
