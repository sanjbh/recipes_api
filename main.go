package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)

	file, err := os.ReadFile("recipes.json")
	if err != nil {
		panic(fmt.Sprintf("Error reading from recipes.json: %s", err.Error()))
	}

	json.Unmarshal([]byte(file), &recipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.Run()
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt"`
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"error": err.Error(),
		})
		return
	}

	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()

	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func ListRecipeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, &recipes)
}

func UpdateRecipeHandler(c *gin.Context) {

	id := c.Param("id")

	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for index, theRecipe := range recipes {
		if theRecipe.ID == id {
			recipes[index] = recipe
			c.JSON(http.StatusOK, recipe)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Recipe not found",
	})
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	for idx, recipe := range recipes {
		if recipe.ID == id {
			recipes = append(recipes[:idx], recipes[idx+1:]...)
			// recipes = slices.Delete(recipes, idx, idx+1)
			c.JSON(http.StatusOK, gin.H{
				"message": "Recipe has been deleted",
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Recipe not found",
	})
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for _, recipe := range recipes {
		for _, t := range recipe.Tags {
			if strings.EqualFold(t, tag) {
				listOfRecipes = append(listOfRecipes, recipe)
				break
			}
		}
	}

	if len(listOfRecipes) > 0 {
		c.JSON(http.StatusOK, listOfRecipes)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No recipes found with the specified tag",
		})
	}
}
