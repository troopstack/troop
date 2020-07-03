package model

type FileRequest struct {
	FileName   string `json:"file_name"`
	File       []byte `json:"file"`
	Dest       string `json:"dest"`
	Cover      bool   `json:"cover"`
	Target     string `json:"target" binding:"required"`
	TargetType string `json:"target_type" default:"server"`
	Tag        string `json:"tag"`
	OS         string `json:"os"`
	Detach     bool   `json:"detach" default:"false"`
	Timeout    int    `json:"timeout"`
}

type ScoutFileRequest struct {
	TaskId   string
	FileName string
	Url      string
	Dest     string
	Cover    bool
}
