package dal

import (
	"log"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	DB = initSqlite()
}

func initSqlite() *gorm.DB {
	var err error
	DB, err := gorm.Open(sqlite.Open("cicd.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("gorm open database error: %s", err)
	}

	if err = DB.AutoMigrate(&Pipeline{}, &Step{}, &Job{}, &JobRunner{}, &Runner{}, &RunnerLabel{}, &Git{}); err != nil {
		panic(err)
	}
	hlog.Info("sqlite init success")
	return DB
}
