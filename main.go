package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WeatherData struct {
	City        string    `bson:"city" json:"name"`
	Description string    `bson:"description" json:"description"`
	Temp        float64   `bson:"temp" json:"temp"`
	LastUpdated time.Time `bson:"last_updated"`
}

type weatherjson struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`

	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`

	Name string `json:"name"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	BASE_URL := os.Getenv("BASE_URL")
	API_KEY := os.Getenv("API_KEY")
	MONGO_URI := os.Getenv("MONGO_URI")

	fmt.Println("Where you want to check the weather:")

	var city string

	fmt.Scanln(&city)

	SEARCH_URL := fmt.Sprintf("%v?appid=%s&q=%s", BASE_URL, API_KEY, city)

	response, err := http.Get(SEARCH_URL)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	//context settings

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGO_URI))
	if err != nil {
		log.Fatal("Failed to create MONGODB Client:", err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			log.Fatal("Failed to disconnect MONGODB:", err)
		}
	}()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	weatherCollection := client.Database("weatherdb").Collection("weather")

	weatherBytes, _ := io.ReadAll(response.Body)
	weather := weatherjson{}
	json.Unmarshal(weatherBytes, &weather)

	//preparing the data for mongodb

	weatherData := WeatherData{
		City:        weather.Name,
		Description: weather.Weather[0].Description,
		Temp:        weather.Main.Temp - 273.15,
		LastUpdated: time.Now(),
	}

	// upsert data into MONGODB

	filter := bson.M{"city": weatherData.City}
	update := bson.M{"$set": weatherData}
	opts := options.Update().SetUpsert(true)

	_, err = weatherCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Fatal("failed to update data:", nil)
	}

	fmt.Printf("Today the weather in %v is %v, temperature is %2.2v.", city, weather.Weather[0].Description, weather.Main.Temp-273.15)

}
