package controller

import (
	"errors"
	"fmt"
	"log"
	"metua/app/model"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/twinj/uuid"
)

var redisClient *redis.Client

func init() {
	//Initializing redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	redisClient = redis.NewClient(&redis.Options{
		Addr: dsn, //redis port
	})
	_, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
}

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
	bearToken := r.Header.Get("Authorization")
	fmt.Println(bearToken)
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexptected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("Access_token")), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return nil
	}

	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}

	return nil
}

func ExtractTokenMetadata(r *http.Request) (*model.AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}

		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		if err != nil {
			return nil, err
		}

		return &model.AccessDetails{
			AccessUuid: accessUuid,
			UserId:     int64(userId),
		}, nil
	}
	return nil, err
}

func FetchAuth(authD *model.AccessDetails) (int64, error) {
	userid, err := redisClient.Get(authD.AccessUuid).Result()
	if err != nil {
		return 0, err
	}

	userID, _ := strconv.ParseUint(userid, 10, 64)
	if authD.UserId != int64(userID) {
		return 0, errors.New("unauthorized")
	}
	return int64(userID), nil
}
func CreateTodo(c *gin.Context) {
	var td model.Todo
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
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	td.UserID = userId

	c.JSON(http.StatusCreated, td)

}

func DeleteAuth(givenUuid string) (int64, error) {
	deleted, err := redisClient.Del(givenUuid).Result()
	if err != nil {
		return 0, nil
	}

	return deleted, nil
}

func Logout(c *gin.Context) {
	au, err := ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	deleted, delErr := DeleteAuth(au.AccessUuid)
	if delErr != nil || deleted == 0 {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}

	c.JSON(http.StatusOK, "Successfully logged out")
}
