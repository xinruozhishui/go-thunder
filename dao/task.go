package dao
import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/xinruozhishui/go-thunder/model"
)

// 创建新任务
func CreateTask(data model.Task) error {
	task := model.Task{
		Id: data.Id,
		FileName: data.FileName,
		Size: data.Size,
		Url: data.Url,
		DownloadProgress: data.DownloadProgress,
	}
	if err := model.DB().Save(&task).Error; err != nil {
		return err
	}
	return nil
}

func GetTaskList() ([]*model.Task, error) {
	var (
		task []*model.Task
	)
	if err := model.DB().Find(&task).Error; err != nil {
		return nil, err
	}
	return task, nil
}

// 更新任务
func UpdateTask(data *model.Task) error {
	if err := model.DB().Model(&model.Task{}).Updates(data).Error; err != nil {
		return err
	}
	return nil
}