package models

import "time"

type AILog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    *int      `json:"user_id"`
	InputText string    `json:"input_text"`
	AIOutput  string    `json:"ai_output"` // کامل متن پاسخ از مدل (طبیعی + system_output)
	CreatedAt time.Time `json:"created_at"`
}
