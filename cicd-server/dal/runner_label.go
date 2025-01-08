package dal

type RunnerLabel struct {
	RunnerID uint   `gorm:"integer"`
	Label    string `gorm:"text"`
}
