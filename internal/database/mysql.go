package database

import (
	"fmt"

	"github.com/verizhang/file-manager/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQLConnection(
	cfg config.DBConfig,
) (*gorm.DB, error) {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.TLS,
	)

	db, err := gorm.Open(
		mysql.Open(dsn),
		&gorm.Config{},
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}
