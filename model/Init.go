package model

import (
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/jinzhu/gorm"
	"fmt"
	"github.com/xinruozhishui/go-thunder/common"
)

var (
	db *gorm.DB
	err error
)

func Init()  {
	db, err = gorm.Open("sqlite3", common.GetDbPath())
	if err != nil {
		fmt.Println("OpenSqliteErr:",err)
	}
	db.AutoMigrate(&Task{})
}

func DB() *gorm.DB  {
	return db
}