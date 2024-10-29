package models

type User struct {
	Id           int64
	Username     string
	PasswordHash []byte
}
