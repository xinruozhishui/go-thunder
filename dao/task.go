package dao
import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/xinruozhishui/go-thunder/model"
)

// 创建新任务
func CreateTask(data map[string]interface{}) error {
	task := model.Task{
		Id: data["id"].(int64),
		FileName: data["file_name"].(string),
		Size: data["size"].(int64),
		Downloaded: data["downloaded"].(int64),
		Progress: data["progress"].(int64),
		Speed: data["speed"].(int64),
	}
	if err := model.DB().Save(&task).Error; err != nil {
		return err
	}
	return nil
}

// 更新任务
func UpdateTask(data map[string]interface{}) error {
	if err := model.DB().Model(&model.Task{}).Updates(data).Error; err != nil {
		return err
	}
	return nil
}