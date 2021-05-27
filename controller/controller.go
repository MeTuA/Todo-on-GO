package controller

import (
	"fmt"
	"log"
	"metua/app/db"
	"metua/app/model"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func ExtractToken(r *http.Request) string {
	tokenString := r.Header.Get("Authorization")

	strArr := strings.Split(tokenString, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	defer r.Body.Close()
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenStr := ExtractToken(r)
	defer r.Body.Close()
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexptected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("Access_token")), nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return token, nil

}

func ExtractTokenMetadata(r *http.Request) (*model.AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Body.Close()
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			log.Fatal(err)
		}

		return &model.AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}

	return nil, err
}
func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		log.Fatal(err)
	}
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		log.Fatal(ok)
	}
	defer r.Body.Close()
	return nil
}

func FetchAuth(authD *model.AccessDetails) (uint64, error) {
	userid, err := redisClient.Get(authD.AccessUuid).Result()
	if err != nil {
		log.Fatal(err)
	}

	userId, _ := strconv.ParseUint(userid, 10, 64)

	defer redisClient.Close()
	return userId, nil
}

func CreateTodo(c *gin.Context) {
	var td *model.Todo
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}

	tokenAuth, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	userId, err := FetchAuth(tokenAuth)
	if err != nil {
		log.Fatal(err)
	}

	td.UserID = int64(userId)

	c.JSON(http.StatusCreated, td)
}
