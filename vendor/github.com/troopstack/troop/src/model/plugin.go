package model

type PluginRequest struct {
	Plugin     string `json:"plugin"`
	Action     string `json:"action" binding:"required"`
	Args       string `json:"args"`
	ConfigByte []byte `json:"config_byte"`
	ConfigName string `json:"config_name"`
	NoCheck    bool   `json:"no_check"  default:"false"`
	Target     string `json:"target" binding:"required"`
	TargetType string `json:"target_type" default:"server"`
	Tag        string `json:"tag"`
	OS         string `json:"os"`
	Detach     bool   `json:"detach" default:"false"`
	Timeout    int    `json:"timeout"`
	Priority   int    `json:"priority"`
}

type ScoutPluginRequest struct {
	Plugin  string
	TaskId  string
	Action  string
	Args    string
	FileUrl string
	Plugins map[string]interface{}
}

type UpdateScoutHavePluginRequest struct {
	Hostname string   `json:"hostname"`
	Plugins  []string `json:"plugins"`
}

type RpcInitDataResponse struct {
	Plugins        map[string]interface{}
	IgnoreCommands []string
}
