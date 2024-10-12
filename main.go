package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Item struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title       string             `json:"title,omitempty" bson:"title,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Price       string             `json:"price,omitempty" bson:"price,omitempty"`
	CreatedAt   string             `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	Category    string           `json:"category,omitempty" bson:"category,omitempty"`
}

var Items *mongo.Collection

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MOGNODB_URI := os.Getenv("MONGODB_URI")
	clientOptions := options.Client().ApplyURI(MOGNODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	Items = client.Database("go-mongo").Collection("items")

	app := fiber.New()

	app.Get("/api/items", getItems)
	app.Post("api/items", createItem)
	app.Get("api/items/:id", getItem)
	app.Patch("api/items/:id", updateItem)
	app.Delete("api/items/:id", deleteTodo)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000"
	}

	log.Fatal(app.Listen(PORT))
}

func getItems(c *fiber.Ctx) error {
	var items []Item

	page := c.Query("page"); if page == "" { page = "1" }

	sortBy := c.Query("sortBy")
	if sortBy == "" {
		sortBy = "createdAt"
	}
	sortOrder := c.Query("sortOrder")
	if sortOrder == "" {
		sortOrder = "desc"
	}
	category := c.Query("category")

	filter := bson.M{}

	if category != "" {
		filter = bson.M{"category": category}
	}

	var sort = bson.D{}

	if sortOrder == "desc" {
		sort = bson.D{{Key: sortBy, Value: -1}}
	} else {
		sort = bson.D{{Key: sortBy, Value: 1}}
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid page number"})
	}

	limit := int64(10 * pageInt)
	log.Println(limit)

	cursor, err := Items.Find(context.Background(), filter, &options.FindOptions{
		Sort: sort,
		Limit: &limit,
	})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var item Item
		if err := cursor.Decode(&item); err != nil {
			return err
		}
		items = append(items, item)
	}

	return c.JSON(items)
}

func getItem(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	cursor := Items.FindOne(context.Background(), bson.M{"_id": objectID})

	item := &Item{}
	if err := cursor.Decode(item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching item"})
	}

	return c.JSON(item)
}

func createItem(c *fiber.Ctx) error {
	item := new(Item)

	if err := c.BodyParser(item); err != nil {
		return err
	}

	if item.Title == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}
	if item.Description == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Description is required"})
	}
	if item.Price == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Price is required"})
	}
	if item.Category == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Category is required"})
	}

	item.CreatedAt = time.Now().GoString()

	insertResult, err := Items.InsertOne(context.Background(), item)

	if err != nil {
		return err
	}

	item.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(item)
}

func updateItem(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	newItem := new(Item)

	if err := c.BodyParser(newItem); err != nil {
		return err
	}

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	cursor := Items.FindOne(context.Background(), bson.M{"_id": objectID})

	item := &Item{}
	if err := cursor.Decode(item); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error fetching item"})
	}

	if (newItem.Title == "") && (newItem.Description == "") && (newItem.Price == "") && (newItem.Category == "") {
		return c.Status(400).JSON(fiber.Map{"error": "Title, Description, Price or Category is required"})
	}

	if newItem.Title != "" {
		item.Title = newItem.Title
	}
	if newItem.Description != "" {
		item.Description = newItem.Description
	}
	if newItem.Price != "" {
		item.Price = newItem.Price
	}
	if newItem.Category != "" {
		item.Category = newItem.Category
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"title": item.Title, "description": item.Description, "price": item.Price, "category": item.Category}}

	_, err = Items.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}
	return c.Status(200).JSON(fiber.Map{"message": "success"})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	filter := bson.M{"_id": objectID}

	_, err = Items.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"message": "success"})
}
