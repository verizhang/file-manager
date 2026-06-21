package mysql

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLConfig struct {
	Host     string `envconfig:"DB_HOST" required:"true"`
	Port     int    `envconfig:"DB_PORT" default:"3306"`
	User     string `envconfig:"DB_USER" required:"true"`
	Password string `envconfig:"DB_PASSWORD"`
	Name     string `envconfig:"DB_NAME" required:"true"`
	TLS      string `envconfig:"DB_TLS" default:"skip-verify"`
}

func NewMySQLConnection(
	cfg MySQLConfig,
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
