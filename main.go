package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"recipes-api/handlers"
	"recipes-api/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// var recipes []models.Recipe
// var collection *mongo.Collection
var recipesHandler handlers.RecipesHandler

func init() {

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatalf("Unable to connect to MongoDB: %s", err.Error())
	}
	log.Println("Connected to MongoDB")

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	recipesHandler = *handlers.NewRecipesHandler(ctx, collection)

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

	recipesHandler = *handlers.NewRecipesHandler(ctx, collection)

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
	file, err := os.ReadFile("recipes.json")
	if err != nil {
		return err
	}

	recipes := make([]models.Recipe, 0)

	if err := json.Unmarshal(file, &recipes); err != nil {
		return err
	}

	listOfRecipes := make([]any, len(recipes))

	for i, r := range recipes {
		listOfRecipes[i] = r
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
