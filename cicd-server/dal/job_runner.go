package dal

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type JobRunner struct {
	gorm.Model
	JobID           uint
	StepID          uint
	StepSort        int
	Commands        ListString
	Status          Status
	EventStatus     EventStatus
	Message         string
	AssignRunnerIds AssignRunnerIds
	Trigger         Trigger
	StartTime       time.Time
	EndTime         time.Time
}

type Status string

const (
	Pending        Status = "pending"
	Assigning      Status = "assigning"
	Running        Status = "running"
	PartialRunning Status = "partial_running"
	Success        Status = "success"
	PartialSuccess Status = "partial_success"
	Failed         Status = "failed"
)

type EventStatus map[Status]int

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *EventStatus) Scan(value interface{}) error {
	val := make(EventStatus)
	if err := json.Unmarshal(value.([]byte), &val); err != nil {
		return err
	}
	*j = val
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j EventStatus) Value() (driver.Value, error) {
	if len(j) == 0 {
		return json.Marshal(map[Status]int{})
	}
	return json.Marshal(j)
}

type AssignRunnerIds []uint

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *AssignRunnerIds) Scan(value interface{}) error {
	val := make(AssignRunnerIds, 0)
	if err := json.Unmarshal(value.([]byte), &val); err != nil {
		return err
	}
	*j = val
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j AssignRunnerIds) Value() (driver.Value, error) {
	if len(j) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(j)
}
