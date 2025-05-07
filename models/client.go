package models

import (
	"time"

	"gorm.io/gorm"
)

type Client struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"size:250;not null"`
	Slug         string    `gorm:"size:100;not null;"`
	IsProject    string    `gorm:"size:30;check:(is_project in ('0','1'));not null;default:'0'"`
	SelfCapture  string    `gorm:"size:1;not null;default:'1'"`
	ClientPrefix string    `gorm:"size:4;not null"`
	ClientLogo   string    `gorm:"size:255;not null;default:'no-image.jpg'"`
	Address      string    `gorm:"type:text"`
	PhoneNumber  string    `gorm:"size:50"`
	City         string    `gorm:"size:50"`
	CreatedAt    time.Time `gorm:"default:null"`
	UpdatedAt    time.Time `gorm:"default:null"`
	DeletedAt    gorm.DeletedAt
}

func (Client) TableName() string {
	return "my_client"
}
