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

type DeathData struct {
	Date   string
	Deaths string
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

	//keys, err := client.Keys("2020*").Result()
	deathData := &DeathData{
		Date:   timeNow,
		Deaths: deaths,
	}

	bytes, err := json.Marshal(deathData)
	if err != nil {
		fmt.Println(err)
	}

	client.LPush("deaths", bytes)

	deathDatas, err := client.LRange("deaths", 0, 10).Result()
	if err != nil {
		fmt.Println(err)
	}

	for _, currDeathData := range deathDatas {
		var currDeathDataParsed DeathData
		json.Unmarshal([]byte(currDeathData), &currDeathDataParsed)
		fmt.Println(string(currDeathDataParsed.Deaths) + " - " + currDeathDataParsed.Date)
	}
}
