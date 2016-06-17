package main

import (
	"encoding/gob"
)

func init() {
	gob.Register(UserState{})
	gob.Register(ExtendedState{})
}

type ListContext struct {
	Filter  string
	Logged bool
	UserInfo User
	Items []Item
}

type Item struct {
	Id, Ke_user_id, MinecraftId, Category, Name string
	Found bool
	I1, I2, I3, I4, I5, I6, I7, I8, I9 int64
	notified_obsolete, notified_wrong, has_image, current bool
	date_created string
}

type User struct {
	Id, Email string
}

type UserState struct {
	Logged     bool
	UserID     int64
	Nick       string
	Email      string
	Lang       string
	RememberMe bool
}
type UserRow struct {
	UserID         int64
	Enabled        bool
	Nick           string
	Email          string
	Password       string
	ActivationCode string
	RecoveryCode   string
	Lang           string
}


type ExtendedState struct {
	UserID    int64
	Lang      string
	LastLogin string
}

type LoginCredentials struct {
	Email, Password string
	RememberMe      bool `sql:"default: false"`
}
