package controller

import (
	"net/http"

	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/service"

	"github.com/alexedwards/argon2id"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LoginPayload ...
type LoginPayload struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

// Login ...
func Login(c *gin.Context) {
	var payload LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	if !service.IsEmailValid(payload.Email) {
		render(c, gin.H{"msg": "wrong email address"}, http.StatusBadRequest)
		return
	}

	v, err := service.GetUserByEmail(payload.Email)
	if err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	verifyPass, err := argon2id.ComparePasswordAndHash(payload.Password, v.Password)
	if err != nil {
		log.WithError(err).Error("error code: 1011")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	if !verifyPass {
		render(c, gin.H{"msg": "wrong credentials"}, http.StatusUnauthorized)
		return
	}

	accessJWT, err := middleware.GetJWT(v.AuthID, v.Email, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1012")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, err := middleware.GetJWT(v.AuthID, v.Email, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1013")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	render(c, jwtPayload, http.StatusOK)
}

// Refresh ...
func Refresh(c *gin.Context) {
	authID := middleware.AuthID
	email := middleware.Email

	// check validity
	if authID == 0 {
		render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}
	if email == "" {
		render(c, gin.H{"msg": "access denied"}, http.StatusUnauthorized)
		return
	}

	// issue new tokens
	accessJWT, err := middleware.GetJWT(authID, email, "access")
	if err != nil {
		log.WithError(err).Error("error code: 1021")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}
	refreshJWT, err := middleware.GetJWT(authID, email, "refresh")
	if err != nil {
		log.WithError(err).Error("error code: 1022")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		return
	}

	jwtPayload := middleware.JWTPayload{}
	jwtPayload.AccessJWT = accessJWT
	jwtPayload.RefreshJWT = refreshJWT
	render(c, jwtPayload, http.StatusOK)
}
