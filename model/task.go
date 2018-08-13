package model

type Task struct {
	Id int64 ` json:"id"`
	FileName string `json:"file_name"`
	Size int64 `json:"size"`
	Downloaded int64 `json:"downloaded"`
	Progress int64 `json:"progress"`
	Speed int64 `json:"speed"`
}