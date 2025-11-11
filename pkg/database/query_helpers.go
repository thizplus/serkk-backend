package database

import (
	"gorm.io/gorm"
)

// QueryHelper provides common query optimizations
type QueryHelper struct {
	db *gorm.DB
}

// NewQueryHelper creates a new query helper
func NewQueryHelper(db *gorm.DB) *QueryHelper {
	return &QueryHelper{db: db}
}

// WithPagination adds pagination to query
func WithPagination(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 10
		}
		if limit > 100 {
			limit = 100 // Max limit
		}
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

// WithPreload preloads specified associations
func WithPreload(associations ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, association := range associations {
			db = db.Preload(association)
		}
		return db
	}
}

// WithOrder adds ordering to query
func WithOrder(order string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if order != "" {
			return db.Order(order)
		}
		return db
	}
}

// WithSearch adds search condition
func WithSearch(column, value string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if value != "" {
			return db.Where(column+" ILIKE ?", "%"+value+"%")
		}
		return db
	}
}

// WithFilter adds filter condition
func WithFilter(column string, value interface{}) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if value != nil {
			return db.Where(column+" = ?", value)
		}
		return db
	}
}

// WithDateRange adds date range filter
func WithDateRange(column, start, end string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if start != "" && end != "" {
			return db.Where(column+" BETWEEN ? AND ?", start, end)
		} else if start != "" {
			return db.Where(column+" >= ?", start)
		} else if end != "" {
			return db.Where(column+" <= ?", end)
		}
		return db
	}
}

// WithJoins adds join clauses
func WithJoins(joins ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, join := range joins {
			db = db.Joins(join)
		}
		return db
	}
}

// WithSelect specifies fields to select
func WithSelect(fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(fields) > 0 {
			return db.Select(fields)
		}
		return db
	}
}

// OptimizeQuery applies common optimizations
func OptimizeQuery(db *gorm.DB) *gorm.DB {
	return db.
		// Use index hints where applicable
		Set("gorm:query_option", "FOR UPDATE SKIP LOCKED")
}

// BatchInsert performs batch insert with optimal chunk size
func BatchInsert(db *gorm.DB, records interface{}, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100 // Default batch size
	}
	return db.CreateInBatches(records, batchSize).Error
}

// CountWithCache counts records with caching capability
func CountWithCache(db *gorm.DB, model interface{}) (int64, error) {
	var count int64
	err := db.Model(model).Count(&count).Error
	return count, err
}

// ExistsCheck checks if record exists (more efficient than Count)
func ExistsCheck(db *gorm.DB, query interface{}, args ...interface{}) (bool, error) {
	var exists bool
	err := db.Model(query).
		Select("1").
		Where(query, args...).
		Limit(1).
		Scan(&exists).Error
	return exists, err
}

// FindWithCache finds record with caching capability
// Note: Actual caching implementation would require a cache layer
func FindWithCache(db *gorm.DB, dest interface{}, cacheKey string) error {
	// Placeholder for cache check
	// if cached := cache.Get(cacheKey); cached != nil {
	//     return json.Unmarshal(cached, dest)
	// }

	err := db.First(dest).Error
	if err != nil {
		return err
	}

	// Placeholder for cache set
	// cache.Set(cacheKey, dest, ttl)

	return nil
}

// BulkUpdate performs bulk update efficiently
func BulkUpdate(db *gorm.DB, model interface{}, updates map[string]interface{}, where interface{}, args ...interface{}) error {
	return db.Model(model).
		Where(where, args...).
		Updates(updates).Error
}

// SoftDeleteBatch soft deletes multiple records efficiently
func SoftDeleteBatch(db *gorm.DB, model interface{}, ids []interface{}) error {
	return db.Where("id IN ?", ids).Delete(model).Error
}
