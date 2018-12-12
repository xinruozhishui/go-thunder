package model

type Task struct {
	Id int64 ` json:"id"`
	FileName string `gorm:"type:varchar(20);default:'';description:''" json:"file_name"`
	Size int64 `gorm:"type:bigint(20);default:0;description:''" json:"size"`
	Url string `gorm:"type:varchar(255);default:'';description:'download_url'" json:"url"`
	DownloadProgress string `gorm:"type:text;default:'';description:''" json:"download_progress"`
}