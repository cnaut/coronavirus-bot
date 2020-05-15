package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type CountryData struct {
	Active             int `json:"active"`
	Cases              int `json:"cases"`
	CasesPerOneMillion int `json:"casesPerOneMillion"`
	Deaths             int `json:"deaths"`
}

func main() {
	resp, err := http.Get("https://corona.lmao.ninja/v2/countries?sort=cases")

	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	var result []CountryData

	json.Unmarshal(body, &result)

	timeNow := time.Now().String()
	deaths := strconv.Itoa(result[0].Deaths)
	fmt.Println(deaths + " deaths in the United States as of " + timeNow)

	// Save result to redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	err = client.Set(timeNow, deaths, 0).Err()
	if err != nil {
		fmt.Println(err)
	}

	keys, err := client.Keys("2020*").Result()
	if err != nil {
		fmt.Println(err)
	}

	for _, key := range keys {
		deaths, _ := client.Get(key).Result()
		fmt.Println(deaths + " - " + key)
	}
}
