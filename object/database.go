package object

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Database *gorm.DB

func init() {
	os.MkdirAll("data", 0755)
	var err error
	Database, err = gorm.Open(sqlite.Open("data/pages.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 自动迁移
	if err := Database.AutoMigrate(&Page{}); err != nil {
		panic(err)
	}
}
