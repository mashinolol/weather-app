package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

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
	//loadin
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	BASE_URL := os.Getenv("BASE_URL")
	API_KEY := os.Getenv("API_KEY")

	fmt.Println("Where you want to check the weather:")

	var city string

	fmt.Scanln(&city)

	SEARCH_URL := fmt.Sprintf("%v?appid=%s&q=%s", BASE_URL, API_KEY, city)

	response, err := http.Get(SEARCH_URL)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		weatherBytes, _ := io.ReadAll(response.Body)
		weather := weatherjson{}
		json.Unmarshal(weatherBytes, &weather)

		fmt.Printf("Today the weather in %v is %v, temperature is %2.2v.", city, weather.Weather[0].Description, weather.Main.Temp-273.15)
	}
}
