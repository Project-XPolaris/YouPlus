package service

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/projectxpolaris/youplus/config"
	"time"
)

var (
	NeedCreateAdminError = errors.New("need create admin")
	PermissionError      = errors.New("permission denied")
	InvalidateUserError  = errors.New("invalidate user or password")
)

func UserLogin(username string, password string) (string, error) {
	user := DefaultUserManager.GetUserByName(username)
	if user == nil {
		return "", UserNotFoundError
	}
	group := DefaultUserManager.GetGroupByName("youplusadmin")
	if group == nil {
		return "", NeedCreateAdminError
	}
	if !group.HasUser(username) {
		return "", PermissionError
	}
	if !DefaultUserManager.CheckPassword(username, password) {
		return "", InvalidateUserError
	}
	// Create the Claims
	claims := &jwt.StandardClaims{
		Id:        username,
		ExpiresAt: time.Now().Add(15 * time.Hour * 24).Unix(),
		Issuer:    "YouPlusService",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(config.Config.ApiKey))
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ParseUser(tokenString string) (*SystemUser, error) {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config.ApiKey), nil
	})
	if err != nil {
		return nil, err
	}
	user := DefaultUserManager.GetUserByName(claims.Id)
	if user == nil {
		return nil, UserNotFoundError
	}
	return user, nil
}
