package dal

import "gorm.io/gorm"

type Git struct {
	gorm.Model
	PipelineID uint
	Repository string
	Branch     string
	Username   string
	Password   string
	CommitID   string
}
