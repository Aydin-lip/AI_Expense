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
	return &PurchaseService{Repo: repo, DB: repo.DB}
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

func (s *PurchaseService) Query(filter models.PurchaseFilter) ([]models.Purchase, error) {
	if s.DB == nil {
		return nil, errors.New("database is not initialized")
	}

	db := s.DB.Model(&models.Purchase{})

	// user_ids
	if len(filter.UserIDs) > 0 {
		db = db.Where("user_id IN ?", filter.UserIDs)
	}

	// categories
	if len(filter.Categories) > 0 {
		db = db.Where("category IN ?", filter.Categories)
	}

	// date range
	if filter.FromDate != nil {
		db = db.Where("purchase_time >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil {
		db = db.Where("purchase_time <= ?", *filter.ToDate)
	}

	// amount range
	if filter.MinAmount != nil {
		db = db.Where("amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		db = db.Where("amount <= ?", *filter.MaxAmount)
	}

	var res []models.Purchase
	if err := db.Order("purchase_time desc").Find(&res).Error; err != nil {
		return nil, err
	}

	return res, nil
}

func (s *PurchaseService) SumAmount(filter models.PurchaseFilter) (float64, error) {
	db := s.DB.Model(&models.Purchase{})
	// same filter chain as Query...
	var total float64
	if err := db.Select("SUM(amount)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (s *PurchaseService) TopCategory(filter models.PurchaseFilter) (string, float64, error) {
	db := s.DB.Model(&models.Purchase{})
	// filter chain

	type Row struct {
		Category string
		Total    float64
	}
	var rows []Row

	db.Select("category, SUM(amount) as total").
		Group("category").
		Order("total DESC").
		Limit(1).
		Scan(&rows)

	if len(rows) == 0 {
		return "", 0, nil
	}

	return rows[0].Category, rows[0].Total, nil
}
