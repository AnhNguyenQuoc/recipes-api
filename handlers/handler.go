package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"log"
	"net/http"
	"recipes-api/models"
	"time"
)

type RecipesHandler struct {
	db          *sql.DB
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(db *sql.DB, ctx context.Context, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		db:          db,
		ctx:         ctx,
		redisClient: redisClient,
	}
}
func (handler RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	recipe.PublishedAt = time.Now()
	_, err := handler.db.Exec("INSERT INTO recipes (name, tags, ingredients, instructions, published_at) VALUES ($1, $2, $3, $4, $5)", recipe.Name, pq.Array(recipe.Tags), pq.Array(recipe.Ingredients), pq.Array(recipe.Instructions), recipe.PublishedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, recipe)
}

func (handler RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	_, err := handler.db.Exec(
		`UPDATE recipes SET name = $2, tags = $3, ingredients = $4, instructions = $5, published_at = $6 where id = $1`,
		id,
		recipe.Name,
		pq.Array(recipe.Tags),
		pq.Array(recipe.Ingredients),
		pq.Array(recipe.Instructions),
		recipe.PublishedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, recipe)
}

func (handler RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	_, err := handler.db.Exec("DELETE FROM recipes where id = $1", id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe Not Found",
		})

		return
	}

	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func (handler RecipesHandler) SearchRecipesHandler(c *gin.Context) {
	tag := []string{c.Query("tag")}
	listOfRecipes := make([]models.Recipe, 0)

	rows, err := handler.db.Query("SELECT * FROM recipes where tags && $1", pq.Array(tag))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var recipe models.Recipe
		rows.Scan(&recipe.ID, &recipe.Name, pq.Array(&recipe.Tags), pq.Array(&recipe.Ingredients), pq.Array(&recipe.Instructions), &recipe.PublishedAt)
		listOfRecipes = append(listOfRecipes, recipe)
	}

	c.JSON(http.StatusOK, listOfRecipes)
}

func (handler RecipesHandler) ListRecipeHandler(c *gin.Context) {
	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()

	if err == redis.Nil {
		log.Println("Request to Database")

		rows, err := handler.db.Query("SELECT * FROM recipes")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		defer rows.Close()
		recipes := make([]models.Recipe, 0)
		for rows.Next() {
			var recipe models.Recipe
			rows.Scan(&recipe.ID, &recipe.Name, pq.Array(&recipe.Tags), pq.Array(&recipe.Ingredients), pq.Array(&recipe.Instructions), &recipe.PublishedAt)
			recipes = append(recipes, recipe)
		}
		data, _ := json.Marshal(recipes)
		err = handler.redisClient.Set(handler.ctx, "recipes", string(data), 0).Err()
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}
