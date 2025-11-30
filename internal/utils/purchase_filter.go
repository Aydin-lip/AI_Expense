package utils

import (
	"example/AI/internal/models"
	"strconv"
	"time"
)

func ConvertAIFiltersToPurchaseFilter(aiFilters map[string]interface{}, request map[string]interface{}) models.PurchaseFilter {
	var pf models.PurchaseFilter

	// user_ids
	if uIds, ok := request["target_users"].([]interface{}); ok {
		for _, id := range uIds {
			switch v := id.(type) {
			case float64:
				pf.UserIDs = append(pf.UserIDs, int(v))
			case int:
				pf.UserIDs = append(pf.UserIDs, v)
			case string:
				if i, err := strconv.Atoi(v); err == nil {
					pf.UserIDs = append(pf.UserIDs, i)
				}
			}
		}
	}

	// categories
	if cats, ok := aiFilters["categories"].([]interface{}); ok {
		for _, c := range cats {
			if s, ok := c.(string); ok && s != "" {
				pf.Categories = append(pf.Categories, s)
			}
		}
	}

	// from_date / to_date
	parseDate := func(v interface{}) *time.Time {
		s, ok := v.(string)
		if !ok || s == "" {
			return nil
		}
		// try "2006-01-02" first
		t, err := time.Parse("2006-01-02", s)
		if err != nil {
			// try RFC3339
			t2, err2 := time.Parse(time.RFC3339, s)
			if err2 != nil {
				return nil
			}
			return &t2
		}
		return &t
	}

	if fd := parseDate(aiFilters["from_date"]); fd != nil {
		pf.FromDate = fd
	}
	if td := parseDate(aiFilters["to_date"]); td != nil {
		pf.ToDate = td
	}

	// min/max amount
	parseFloat := func(v interface{}) *float64 {
		var f float64
		switch val := v.(type) {
		case float64:
			f = val
		case float32:
			f = float64(val)
		case int:
			f = float64(val)
		case int64:
			f = float64(val)
		case string:
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				f = v
			} else {
				return nil
			}
		default:
			return nil
		}
		if f == 0 {
			return nil // اگر 0 بود، نادیده گرفته بشه
		}
		return &f
	}

	pf.MinAmount = parseFloat(aiFilters["min_amount"])
	pf.MaxAmount = parseFloat(aiFilters["max_amount"])

	return pf
}
