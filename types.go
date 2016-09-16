package main

import (
	"encoding/gob"
	"net/http"
)

func init() {
	gob.Register(UserState{})
	gob.Register(ExtendedState{})
}

type ListContext struct {
	Filter   string
	Logged   bool
	UserInfo User
	Items    []Item
	UserState
}

type Context struct {
	Request *http.Request

	UserState
	ExtendedState

	Flashes []string
}

type Item struct {
	Id, Ke_user_id, MinecraftId, Category, Name           string
	Found                                                 bool
	I1, I2, I3, I4, I5, I6, I7, I8, I9                    int64
	notified_obsolete, notified_wrong, has_image, current bool
	date_created                                          string
}

type ChangePasswordCredentials struct {
	OldPassword, NewPassword, NewPasswordConfirmation string
}

type ResetPasswordCredentials struct {
	NewPassword, NewPasswordConfirmation string
	RecoveryCode string
}

type ResetRequestCredentials struct {
	Email string
}
type ResetPasswordContext struct {
	Context
	RecoveryCode string

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
}

type ExtendedState struct {
	UserID    int64
	LastLogin string
}

type LoginCredentials struct {
	Email, Password string
	RememberMe      bool `sql:"default: false"`
}
type EmailContext struct {
	Request *http.Request

	Nick, Email, ActivationCode string
}
type SignUpCredentials struct {
	Nick, Email, Password, PasswordConfirmation string
}