package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"recipes-api/handlers"
	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// var recipes []models.Recipe
// var collection *mongo.Collection
var recipesHandler *handlers.RecipesHandler

func init() {

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatalf("Unable to connect to MongoDB: %s", err.Error())
	}
	log.Println("Connected to MongoDB")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // or "redis:6379" in Docker
		Password: "",               // set if your Redis requires auth
		DB:       0,                // use default DB
	})
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)

	count, err := GetNumberOfRecordsFromMongoCollection()
	if err != nil {
		log.Fatalf("Error getting collection count from DB: %v\n", err)
	}

	if count == 0 {
		if err := InsertRecipesFromFile(); err != nil {
			log.Fatalf("Error inserting data from file to DB: %v\n", err)
		}
	} else {
		log.Println("Skipping data insert into DB as it is already populated")
	}

	status := redisClient.Ping(ctx)
	fmt.Println(status)

}

func main() {
	router := gin.Default()
	router.POST("/recipes", recipesHandler.NewRecipeHandler)
	router.GET("/recipes", recipesHandler.ListRecipesHandler)
	router.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
	router.GET("/recipes/search", recipesHandler.GetOneRecipeHandler)
	router.Run()
}

func InsertRecipesFromFile() error {

	var tempRecipes []struct {
		ID           string    `json:"id"`
		Name         string    `json:"name" bson:"name"`
		Tags         []string  `json:"tags" bson:"tags"`
		Ingredients  []string  `json:"ingredients" bson:"ingredients"`
		Instructions []string  `json:"instructions" bson:"instructions"`
		PublishedAt  time.Time `json:"publishedAt" bson:"publishedAt"`
	}

	file, err := os.ReadFile("recipes.json")
	if err != nil {
		return err
	}

	if err := json.Unmarshal(file, &tempRecipes); err != nil {
		return err
	}

	// recipes := make([]models.Recipe, len(tempRecipes))
	listOfRecipes := make([]any, len(tempRecipes))

	for index, recipe := range tempRecipes {
		objectId, err := primitive.ObjectIDFromHex(recipe.ID)
		if err != nil {
			objectId = primitive.NewObjectID()
		}
		listOfRecipes[index] = models.Recipe{
			ID:           objectId,
			Name:         recipe.Name,
			Tags:         recipe.Tags,
			Ingredients:  recipe.Ingredients,
			Instructions: recipe.Instructions,
			PublishedAt:  recipe.PublishedAt,
		}
	}

	collection := recipesHandler.GetCollection()

	if _, err := collection.InsertMany(context.Background(), listOfRecipes); err != nil {
		return err
	}

	log.Printf("Inserted recipes: %d\n", len(listOfRecipes))

	return nil
}

func GetNumberOfRecordsFromMongoCollection() (int64, error) {
	collection := recipesHandler.GetCollection()
	count, err := collection.EstimatedDocumentCount(context.Background())
	if err != nil {
		return 0, err
	}

	return count, nil
}
