package controller

import (
	"log"
	"metua/app/db"
	"metua/app/model"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/twinj/uuid"
)

var redisClient = db.Init()

var user = model.User{
	Id:       1,
	Username: "username",
	Password: "password",
}

func Login(c *gin.Context) {
	var u model.User
	if err := c.BindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "Enter correct data")
		return
	}

	if user.Username != u.Username || user.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "Enter correct login or password")
		return
	}

	ts, err := CreateToken(user.Id)
	if err != nil {
		log.Fatal(err)
	}

	err = CreateAuth(user.Id, ts)
	if err != nil {
		log.Fatal(err)
	}

	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}

	c.JSON(http.StatusOK, tokens)

}

func CreateToken(id int) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}

	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AtUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RtUuid = uuid.NewV4().String()

	//Creating access token
	var err error
	os.Setenv("Access_token", "AccessSecret")
	atClaims := &jwt.MapClaims{
		"authorized": true,
		"acess_uuid": td.AtUuid,
		"exp":        td.AtExpires,
		"user_id":    id,
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("Access_token")))
	if err != nil {
		log.Fatal(err)
	}

	//Creating refresh token
	os.Setenv("Refresh_token", "RefreshSecret")
	rtClaims := &jwt.MapClaims{
		"refresh_uuid": td.RtUuid,
		"exp":          td.RtExpires,
		"user_id":      id,
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("Refresh_token")))
	if err != nil {
		log.Fatal(err)
	}

	return td, nil
}

func CreateAuth(id int, td *model.TokenDetails) error {
	at := time.Unix(td.AtExpires, 0)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := redisClient.Set(td.AtUuid, id, at.Sub(now)).Err()
	if errAccess != nil {
		log.Fatal(errAccess)
	}

	errRefresh := redisClient.Set(td.RtUuid, id, rt.Sub(now)).Err()
	if errRefresh != nil {
		log.Fatal(errRefresh)
	}

	return nil
}
