package main

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

