package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/cnaut/cuomo-briefing-tracker/twitter"
)

type geoCountryCode struct {
	Alpha2 string `json:"alpha-2"`
	Name   string `json:"name"`
}
1
type parameters struct {
	GeoCountryCode geoCountryCode `json:"geo-country-code"`
}

type intent struct {
	DisplayName string `json:"displayName"`
}

type queryResult struct {
	Intent     intent     `json:"intent"`
	Parameters parameters `json:"parameters"`
}

type text struct {
	Text []string `json:"text"`
}

type message struct {
	Text text `json:"text"`
}

type countryData struct {
	Active             int `json:"active"`
	Cases              int `json:"cases"`
	CasesPerOneMillion int `json:"casesPerOneMillion"`
	Deaths             int `json:"deaths"`
}

// webhookRequest is used to unmarshal a WebhookRequest JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by
// the Dialogflow protocol buffers:
// https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#WebhookRequest
type webhookRequest struct {
	Session     string      `json:"session"`
	ResponseID  string      `json:"responseId"`
	QueryResult queryResult `json:"queryResult"`
}

// webhookResponse is used to marshal a WebhookResponse JSON object. Note that
// not all members need to be defined--just those that you need to process.
// As an alternative, you could use the types provided by
// the Dialogflow protocol buffers:
// https://godoc.org/google.golang.org/genproto/googleapis/cloud/dialogflow/v2#WebhookResponse
type webhookResponse struct {
	FulfillmentMessages []message `json:"fulfillmentMessages"`
}

// welcome creates a response for the welcome intent.
func welcome(request webhookRequest) (webhookResponse, error) {
	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{"Welcome from Dialogflow Go Webhook"},
				},
			},
		},
	}
	return response, nil
}

// getAgentName creates a response for the get-agent-name intent.
func getAgentName(request webhookRequest) (webhookResponse, error) {
	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{"My name is Dialogflow Go Webhook"},
				},
			},
		},
	}
	return response, nil
}

func handleCoronavirusDataRequest(request webhookRequest) (webhookResponse, error) {
	countryCode := request.QueryResult.Parameters.GeoCountryCode
	countryAlpha2 := countryCode.Alpha2
	countryName := countryCode.Name

	resp, err := http.Get("https://corona.lmao.ninja/v2/countries/" + countryAlpha2)
	if err != nil {
		fmt.Println(err)
		return webhookResponse{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return webhookResponse{}, err
	}

	var result countryData

	json.Unmarshal(body, &result)
	deaths := strconv.Itoa(result.Deaths)

	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{deaths + " deaths in " + countryName},
				},
			},
		},
	}

	return response, nil
}

func handleCuomoBriefingTimeRequest(request webhookRequest) (webhookResponse, error) {
	briefingTime := twitter.FindCuomoBriefingTime()

	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{" Next Cuomo daily briefing is at " + briefingTime},
				},
			},
		},
	}

	return response, nil
}

// handleError handles internal errors.
func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "ERROR: %v", err)
}

// HandleWebhookRequest handles WebhookRequest and sends the WebhookResponse.
func HandleWebhookRequest(w http.ResponseWriter, r *http.Request) {
	var request webhookRequest
	var response webhookResponse
	var err error

	// Read input JSON
	if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Request: %+v", request)

	// Call intent handler
	switch intent := request.QueryResult.Intent.DisplayName; intent {
	case "Default Welcome Intent":
		response, err = welcome(request)
	case "get-agent-name":
		response, err = getAgentName(request)
	case "Coronavirus Data":
		response, err = handleCoronavirusDataRequest(request)
	case "Cuomo Briefing Time":
		response, err = handleCuomoBriefingTimeRequest(request)
	default:
		err = fmt.Errorf("Unknown intent: %s", intent)
	}
	if err != nil {
		handleError(w, err)
		return
	}
	log.Printf("Response: %+v", response)

	// Send response
	if err = json.NewEncoder(w).Encode(&response); err != nil {
		handleError(w, err)
		return
	}
}
