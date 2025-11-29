package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"example/AI/internal/models"
	"example/AI/internal/store"

	"gorm.io/gorm"
)

type PurchaseService struct {
	Repo *store.PurchaseRepo
	DB   *gorm.DB // یا مستقیم *gorm.DB اگر داری
}

func NewPurchaseService(repo *store.PurchaseRepo) *PurchaseService {
	return &PurchaseService{Repo: repo}
}

func toFloat64(v interface{}) (float64, error) {
	switch t := v.(type) {
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		// try parse
		return strconv.ParseFloat(t, 64)
	default:
		return 0, fmt.Errorf("unsupported type for amount: %T", v)
	}
}

func (s *PurchaseService) CreateFromAIData(userID int, aiData map[string]interface{}) (*models.Purchase, error) {
	// Map fields with safe conversions and defaults
	title := ""
	if v, ok := aiData["title"].(string); ok {
		title = v
	}
	var amount float64
	if v, ok := aiData["amount"]; ok {
		if f, err := toFloat64(v); err == nil {
			amount = f
		}
	}
	currency := ""
	if v, ok := aiData["currency"].(string); ok {
		currency = v
	}
	category := ""
	if v, ok := aiData["category"].(string); ok {
		category = v
	}
	subcategory := ""
	if v, ok := aiData["subcategory"].(string); ok {
		subcategory = v
	}
	var vendor *string
	if v, ok := aiData["vendor"].(string); ok && v != "" {
		vendor = &v
	}
	var ptime *time.Time
	if v, ok := aiData["purchase_time"].(string); ok && v != "" {
		// try parse common formats (ISO)
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			ptime = &t
		} else {
			// try date only
			if t2, err2 := time.Parse("2006-01-02", v); err2 == nil {
				ptime = &t2
			}
		}
	}
	if ptime == nil {
		now := time.Now().UTC()
		ptime = &now
	}
	necessity := ""
	if v, ok := aiData["necessity"].(string); ok {
		necessity = v
	}
	emotionalTone := ""
	if v, ok := aiData["emotional_tone"].(string); ok {
		emotionalTone = v
	}
	reasonGuess := ""
	if v, ok := aiData["reason_guess"].(string); ok {
		reasonGuess = v
	}
	confidence := 0.0
	if v, ok := aiData["confidence"]; ok {
		if f, err := toFloat64(v); err == nil {
			confidence = f
		}
	}

	p := &models.Purchase{
		UserID:        userID,
		Title:         title,
		Amount:        amount,
		Currency:      currency,
		Category:      category,
		Subcategory:   subcategory,
		Vendor:        vendor,
		PurchaseTime:  ptime,
		CreatedAt:     time.Now().UTC(),
		Necessity:     necessity,
		EmotionalTone: emotionalTone,
		ReasonGuess:   reasonGuess,
		Confidence:    confidence,
		Status:        "confirmed",
	}

	// basic validation
	if p.Title == "" || p.Amount <= 0 {
		return nil, errors.New("invalid purchase: missing title or amount")
	}

	if err := s.Repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}
