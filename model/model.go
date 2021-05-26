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
