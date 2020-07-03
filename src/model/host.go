package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Host struct {
	gorm.Model
	Hostname     string    `gorm:"type:varchar(512);not null"`
	Ip           string    `gorm:"type:varchar(64)"`
	OS           string    `gorm:"type:varchar(64)"`
	Tags         []Tag     `gorm:"many2many:host_tag;"`
	Plugins      string    `gorm:"type:varchar(1024)"`
	ScoutVersion string    `gorm:"type:varchar(64);column:scout_version"`
	Type         string    `gorm:"type:varchar(64);not null;default:'server'"`
	Status       string    `gorm:"type:varchar(64);not null"`
	AES          string    `gorm:"type:varchar(128)"`
	HandshakeAt  time.Time `gorm:"column:handshake_at"`
}

type Tag struct {
	ID    int
	Name  string `gorm:"type:varchar(256);unique_index"`
	Hosts []Host `gorm:"many2many:host_tag;"`
}
