// Recipes API
//
// This is a sample recipes API. You can find out more about
//	the API at https://github.com/PacktPublishing/BuildingDistributed-Applications-in-Gin.
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Kai
// <mohamed@labouardy.com> https://labouardy.com
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"context"
	"database/sql"
	"github.com/gin-contrib/sessions"
	redisSession "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"recipes-api/database"
	"recipes-api/handlers"
)

var recipesHandler *handlers.RecipesHandler
var authHandler *handlers.AuthHandler
var redisClient *redis.Client
var db *sql.DB
var err error
var ctx context.Context

func init() {
	ctx = context.Background()
	db, err = database.NewDB()
	if err != nil {
		log.Fatal(err)
	}
	redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379", Password: "", DB: 0})
	recipesHandler = handlers.NewRecipesHandler(db, ctx, redisClient)
	authHandler = handlers.NewAuthHandler(ctx, db)
	//initUsers()
}

func main() {
	router := gin.Default()
	store, _ := redisSession.NewStore(10, "tcp", "localhost:6379", "", []byte("secret"))
	router.Use(sessions.Sessions("recipes_api", store))
	router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
	router.GET("/recipes", recipesHandler.ListRecipeHandler)
	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/signout", authHandler.SignOutHandler)

	authorized := router.Group("/", AuthMiddleware())

	authorized.POST("/refresh", authHandler.RefreshHandler)
	authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
	authorized.PUT("/recipe/:id", recipesHandler.UpdateRecipeHandler)
	authorized.DELETE("recipe/:id", recipesHandler.DeleteRecipeHandler)

	err := router.Run()
	if err != nil {
		return
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionToken := session.Get("token")

		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Not Logged"})
			c.Abort()
		}

		c.Next()
		//tokenValue := c.GetHeader("Authorization")
		//claims := &handlers.Claims{}
		//
		//tkn, err := jwt.ParseWithClaims(tokenValue, claims, func(token *jwt.Token) (interface{}, error) {
		//	return []byte("secretkey"), nil
		//})
		//
		//if err != nil {
		//	log.Println(err.Error())
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//}
		//
		//if tkn == nil || !tkn.Valid {
		//	log.Println(tkn)
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//}
		//c.Next()
	}
}

//func initUsers() {
//	users := map[string]string{
//		"admin":      "fCRmh4Q2J7Rseqkz",
//		"packt":      "RE4zfHB35VPtTkbT",
//		"mlabouardy": "L3nSFRcZzNQ67bcc",
//	}
//
//	for username, password := range users {
//		tmp := sha256.Sum256([]byte(password))
//		_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, hex.EncodeToString(tmp[:]))
//		if err != nil {
//			log.Println(err.Error())
//		}
//	}
//}
