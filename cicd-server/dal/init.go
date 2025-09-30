package dal

import (
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	DB = initSqlite()

	initUser()
}

func initUser() {
	user := User{
		Username: "admin",
		Password: "bizyair@123",
		Nickname: "admin",
	}
	username := os.Getenv("CICD_ADMIN_USERNAME")
	if username != "" {
		user.Username = username
	}
	password := os.Getenv("CICD_ADMIN_PASSWORD")
	if password != "" {
		user.Password = password
	}
	if err := user.CreateAdmin(); err != nil {
		log.Fatalf("create user error: %s", err)
	}
}

func initSqlite() *gorm.DB {
	var err error
	DB, err := gorm.Open(sqlite.Open("cicd.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("gorm open database error: %s", err)
	}

	if err = DB.AutoMigrate(
		&Pipeline{},
		&Step{},
		&Job{},
		&JobRunner{},
		&Runner{},
		&RunnerLabel{},
		&Git{},
		&User{},
		&Role{},
		&UserRole{},
		&PipelineRole{},
	); err != nil {
		panic(err)
	}

	hlog.Info("sqlite init success")
	return DB
}
