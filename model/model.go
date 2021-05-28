package model

type User struct {
	Id       int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AtUuid       string
	RtUuid       string
	AtExpires    int64
	RtExpires    int64
}

type Todo struct {
	UserID int64  `json:"user_id"`
	Title  string `json:"title"`
}

type AccessDetails struct {
	AccessUuid string
	UserId     int64
}
