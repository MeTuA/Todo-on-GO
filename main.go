package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var router = gin.Default()

func main() {
	router.POST("/login", Login)
	log.Fatal(router.Run(":8040"))
}

type User struct {
	Id       int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var user = User{
	Id:       1,
	Username: "username",
	Password: "password",
}

func Login(c *gin.Context) {
	var u User
	if err := c.BindJSON(&u); err != nil {
		log.Fatal(err)
		return
	}

	if u.Username != "username" || u.Password != "password" {
		c.JSON(http.StatusUnauthorized, "Enter right info")
		return
	}

	token := CreateToken(user.Id)

	c.JSON(http.StatusOK, token)
}

func CreateToken(id int) string {
	os.Setenv("Access_token", "HEY123")
	atClaims := jwt.MapClaims{
		"authorized": true,
		"exp":        time.Now().Add(time.Minute * 15).Unix(),
	}

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("Access_token")))
	if err != nil {
		log.Fatal(err)
	}

	return token
}
