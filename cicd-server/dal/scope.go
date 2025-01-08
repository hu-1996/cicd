package dal

import "gorm.io/gorm"

func Paginate(pageIndex, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pageIndex <= 0 || pageSize <= 0 {
			return db
		}
		offset := (pageIndex - 1) * pageSize
		if offset < 0 {
			offset = 0
		}
		return db.Offset(offset).Limit(pageSize)
	}
}
