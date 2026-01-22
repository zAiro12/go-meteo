package main

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func InitMongo() {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI non impostata nelle variabili d'ambiente")
	}
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal("Errore connessione MongoDB:", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB non raggiungibile:", err)
	}
	MongoClient = client
	MongoDB = client.Database("meteo")
	log.Println("MongoDB connesso!")
	// try load default global location from DB
	loadDefaultLocationFromDB()
}

// loadDefaultLocationFromDB legge la location di default (se presente) dalla collection app_config
func loadDefaultLocationFromDB() {
	coll := MongoDB.Collection("app_config")
	ctx := context.Background()
	var doc struct {
		Key   string                 `bson:"key"`
		Value map[string]interface{} `bson:"value"`
	}
	if err := coll.FindOne(ctx, bson.M{"key": "default_location"}).Decode(&doc); err != nil {
		return
	}
	latVal, okLat := doc.Value["lat"].(float64)
	lonVal, okLon := doc.Value["lon"].(float64)
	if okLat && okLon {
		locationMutex.Lock()
		customLat = latVal
		customLon = lonVal
		useCustom = true
		locationMutex.Unlock()
		log.Printf("Caricata posizione globale da DB: %.4f, %.4f", customLat, customLon)
	}
}
