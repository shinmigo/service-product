package utils

type ResponseList struct {
	RunTime float64     `json:"run_time"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
