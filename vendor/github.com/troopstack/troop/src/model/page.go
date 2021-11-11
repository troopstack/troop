package model

type PageResponseBody struct {
	Count	int				`json:"count"`
	Data	interface{} 	`json:"data"`
}
