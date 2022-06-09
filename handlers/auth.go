package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"log"
	"net/http"
	"recipes-api/models"
	"time"
)

type AuthHandler struct {
	db  *sql.DB
	ctx context.Context
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JWTOutput struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

func NewAuthHandler(ctx context.Context, db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db:  db,
		ctx: ctx,
	}
}

func (handler *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}

	h := sha256.Sum256([]byte(user.Password))
	err := handler.db.QueryRow("SELECT * FROM users WHERE username = $1 AND password = $2", user.Username, hex.EncodeToString(h[:])).Scan(&user)
	log.Println(err)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	} else if err == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	sessionToken := xid.New().String()
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionToken)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
	//expirationTime := time.Now().Add(10 * time.Minute)
	//claims := &Claims{
	//	Username: user.Username,
	//	StandardClaims: jwt.StandardClaims{
	//		ExpiresAt: expirationTime.Unix(),
	//	},
	//}
	//
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//tokenString, err := token.SignedString([]byte("secretkey"))
	//
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{
	//		"error": err.Error(),
	//	})
	//}
	//
	//jwtOutput := JWTOutput{
	//	Token:  tokenString,
	//	Expire: expirationTime,
	//}
	//
	//c.JSON(http.StatusOK, jwtOutput)
}

func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"message": "Signed out...",
	})
}

func (handler *AuthHandler) RefreshHandler(c *gin.Context) {
	tokenValue := c.GetHeader("Authorization")
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
	}

	if tkn == nil || !tkn.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Token is not expired yet.",
		})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString("secret")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	jwtOutput := JWTOutput{
		Token:  tokenString,
		Expire: expirationTime,
	}

	c.JSON(http.StatusOK, jwtOutput)
}
