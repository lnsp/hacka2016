package main

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	ACCOUNT_PROFILE_RELATION       = "User"
	SQL_FIND_ACCOUNT_BY_TOKEN      = "Token = ?"
	SQL_FIND_ACCOUNT_BY_DEVICE     = "Device = ?"
	SQL_FIND_PROFILE_BY_ID         = "ID = ?"
	SQL_FIND_FRIENDSHIPS_BY_SOURCE = "Source = ?"
	SQL_FIND_FRIENDSHIP            = "Source = ? AND Target = ?"
	SQL_FIND_LATEST                = "Date > ?"
	SQL_DATE_DESC_ORDER            = "date desc"
	SQL_FIND_SESSION_ID            = "Session = ?"
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
	Points    uint
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
