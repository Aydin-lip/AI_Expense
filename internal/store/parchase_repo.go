package store

import (
	"example/AI/internal/models"
	"time"

	"gorm.io/gorm"
)

type PurchaseRepo struct {
	DB *gorm.DB
}

func NewPurchaseRepo(db *gorm.DB) *PurchaseRepo {
	return &PurchaseRepo{DB: db}
}

func (r *PurchaseRepo) Create(p *models.Purchase) error {
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	return r.DB.Create(p).Error
}

// قبلاً QueryPurchasesDynamic که داشتیم اینجا قرار میگیرد (خلاصه نشون نمیدم چون قبلاً فرستادم)
