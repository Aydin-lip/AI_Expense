package models

import "time"

type Purchase struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	UserID        int        `json:"user_id"`
	Title         string     `json:"title"`
	Amount        float64    `json:"amount"`
	Currency      string     `json:"currency"`
	Category      string     `json:"category"`
	Subcategory   string     `json:"subcategory"`
	Vendor        *string    `json:"vendor"`
	PurchaseTime  *time.Time `json:"purchase_time"`
	CreatedAt     time.Time  `json:"created_at"`
	Necessity     string     `json:"necessity"`      // ضروری؟ غیرضروری؟
	EmotionalTone string     `json:"emotional_tone"` // آرام؟ هیجانی؟ عصبی؟
	ReasonGuess   string     `json:"reason_guess"`   // حدس دلیل خرید
	Confidence    float64    `json:"confidence"`     // اعتماد AI
	Status        string     `json:"status"`         // مثلا: "confirmed", "guessed"
}
