package main

import (
	"fmt"
	"time"

	mysql "go.elastic.co/apm/module/apmgormv2/v2/driver/mysql"
	"gorm.io/gorm"
)

type UrlEntity struct {
	gorm.Model
	ID        uint64    `gorm:"primaryKey;column:id"`
	ShortUrl  string    `gorm:"uniqueIndex;column:shour_url"`
	OriginUrl string    `gorm:"column:origin_url"`
	CreateAt  time.Time `gorm:"autoCreateTime;column:create_at"`
	UpdateAt  time.Time `gorm:"autoUpdateTime;column:update_at"`
}

func (UrlEntity) TableName() string {
	return "urls"
}

func InitDB(host, user, password, database string, port int) (db *gorm.DB, err error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, database)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&UrlEntity{})
	return db, nil
}
