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
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"log"
	"recipes-api/database"
	"recipes-api/handlers"
)

var recipesHandler *handlers.RecipesHandler
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
}

func main() {
	router := gin.Default()
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.PUT("/recipe/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("recipe/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.SearchRecipesHandler)
	router.GET("/recipes", recipesHandler.ListRecipeHandler)

	err := router.Run()
	if err != nil {
		return
	}
}
