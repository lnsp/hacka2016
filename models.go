package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Position struct {
	Date      time.Time
	Latitude  float64
	Longitude float64
	Source    uint
}

type Friendship struct {
	gorm.Model
	Source uint
	Target uint
}

// The internal profile model
type Profile struct {
	gorm.Model
	Name      string
	Points    int
	Picture   string
	AccountID uint `gorm:"index"`
	Color     string
}

// The internal account model
type Account struct {
	gorm.Model
	Device string
	Token  string
	User   Profile
}

// The hotspots.
type Hotspot struct {
	gorm.Model
	Session     string
	LastCapture time.Time
	Conqueror   uint
	Token       string
}
