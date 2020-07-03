package model

type TaskScoutInfo struct {
	TaskId    string
	Scout     string
	ScoutType string
	Result    string
	Error     string
	Status    string
}

type TaskResultRequest struct {
	TaskId    string
	Scout     string
	ScoutType string
	Result    string
	Error     string
	Complete  bool
}

type TaskAcceptRequest struct {
	TaskId    string
	Scout     string
	ScoutType string
}

type Env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Task struct {
	Name string `json:"name" binding:"required"`
	Envs []Env  `json:"envs"`
	Dir  string `json:"dir"`
	Args string `json:"args" binding:"required"`
}

type TaskRequest struct {
	Task       []Task `json:"task" binding:"required"`
	Target     string `json:"target" binding:"required"`
	TargetType string `json:"target_type" default:"server"`
	Tag        string `json:"tag"`
	OS         string `json:"os"`
	Detach     bool   `json:"detach" default:"false"`
	Timeout    int    `json:"timeout"`
}

type ScoutTaskRequest struct {
	TaskId string
	Task   []Task
}

type TaskInfoRequest struct {
	TaskId string `json:"task_id" form:"task_id" binding:"required"`
	Wait   bool   `json:"wait" form:"wait" default:"false"`
}
