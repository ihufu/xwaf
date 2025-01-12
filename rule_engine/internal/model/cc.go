package model

import "time"

// CCRule CC防护规则
type CCRule struct {
	ID         int64     `json:"id" db:"id"`
	URI        string    `json:"uri" db:"uri"`
	LimitRate  int       `json:"limit_rate" db:"limit_rate"`
	TimeWindow int       `json:"time_window" db:"time_window"`
	LimitUnit  string    `json:"limit_unit" db:"limit_unit"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// CCRuleQuery CC规则查询条件
type CCRuleQuery struct {
	URI       string `json:"uri"`
	Status    string `json:"status"`
	LimitUnit string `json:"limit_unit"`
}
