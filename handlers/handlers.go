package handlers

import (
	"context"
	"fmt"
	"net/http"
	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection) *RecipesHandler {
	return &RecipesHandler{
		collection,
		ctx,
	}
}

func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	cur, err := handler.collection.Find(handler.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(handler.ctx)

	recipes := make([]models.Recipe, 0)

	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		if err := cur.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, &recipes)

}

func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Unable to parse POST request for inserting new record: %v", err.Error()),
		})
	}

	if _, err := handler.collection.InsertOne(handler.ctx, &recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Unable to insert new record: %v", err.Error()),
		})
	}

	c.JSON(http.StatusOK, &recipe)

}

func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {

	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	/* for index, theRecipe := range recipes {
		if theRecipe.ID == objectId {
			recipes[index] = recipe
			c.JSON(http.StatusOK, recipe)
			return
		}
	} */

	_, err := handler.collection.UpdateOne(context.Background(), bson.M{
		"_id": objectId,
	}, bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "instructions", Value: recipe.Instructions},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "tags", Value: recipe.Tags},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Error while updating record: %v", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been updated",
	})
}

func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	if _, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	})

	var recipe models.Recipe

	if err := cur.Decode(&recipe); err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("Unable to decode the search result: %v", err))
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func (handler *RecipesHandler) GetCollection() *mongo.Collection {
	return handler.collection
}

func (handler *RecipesHandler) GetContext() context.Context {
	return handler.ctx
}
