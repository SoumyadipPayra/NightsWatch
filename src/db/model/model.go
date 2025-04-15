package model

import (
	"time"
)

type User struct {
	ID            uint64    `gorm:"primary_key;auto_increment"`
	UserName      string    `gorm:"unique; index; not null"`
	Password      string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"auto_create_time, default:CURRENT_TIMESTAMP"`
	LastLoginTime time.Time `gorm:"auto_update_time"`
}

func (u *User) TableName() string {
	return "users"
}

type App struct {
	Name    string `gorm:"type:string"`
	Version string `gorm:"type:string"`
}

type DeviceData struct {
	ID             uint64    `gorm:"primary_key;auto_increment"`
	UserID         uint64    `gorm:"not null;index;foreignKey:ID;references:users"`
	InstalledApps  []*App    `gorm:"serializer:json"`
	OSQueryVersion string    `gorm:"type:string"`
	OSVersion      string    `gorm:"type:string"`
	Timestamp      time.Time `gorm:"index;default:CURRENT_TIMESTAMP"`
}

func (d *DeviceData) TableName() string {
	return "device_data"
}
