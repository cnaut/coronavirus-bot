package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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

	fmt.Println(strconv.Itoa(result[0].Deaths) + " deaths in the United States as of " + time.Now().String())
}
