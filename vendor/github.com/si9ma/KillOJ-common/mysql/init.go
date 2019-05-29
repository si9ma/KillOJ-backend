package mysql

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	otgorm "github.com/smacker/opentracing-gorm"
)

func InitDB(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open("mysql", cfg.ConnectionStr)
	if err != nil {
		return nil, err
	}
	// register callbacks must be called for a root instance of your gorm.DB
	otgorm.AddGormCallbacks(db)
	return db, nil
}
