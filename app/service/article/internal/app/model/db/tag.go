package db

import "time"

type Tag struct {
	Id        int64 `gorm:"primary_key";json:"_"`
	Name      string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at" gorm:"-"` //gorm ignore
}

