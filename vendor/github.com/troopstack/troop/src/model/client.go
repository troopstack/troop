package model

type DetachTaskResponse struct {
	TaskId string `json:"task_id"`
	Result string `json:"result"`
	Error  string `json:"error"`
	Code   int    `json:"code"`
}

type CommonResponse struct {
	Result string `json:"result"`
	Error  string `json:"error"`
	Code   int    `json:"code"`
}

type TaskResponse struct {
	TaskId string                    `json:"task_id"`
	Result map[string]*TaskScoutInfo `json:"result"`
	Error  string                    `json:"error"`
	Code   int                       `json:"code"`
}
