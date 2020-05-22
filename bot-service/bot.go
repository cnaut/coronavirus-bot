package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type intent struct {
	DisplayName string `json:"displayName"`
}

type queryResult struct {
	Intent intent `json:"intent"`
}

type text struct {
	Text []string `json:"text"`
}

type message struct {
	Text text `json:"text"`
}

type CountryData struct {
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
	resp, err := http.Get("https://corona.lmao.ninja/v2/countries?sort=cases")
	if err != nil {
		fmt.Println(err)
		return webhookResponse{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return webhookResponse{}, err
	}

	var result []CountryData

	json.Unmarshal(body, &result)
	deaths := strconv.Itoa(result[0].Deaths)

	response := webhookResponse{
		FulfillmentMessages: []message{
			{
				Text: text{
					Text: []string{deaths + " deaths in the United States"},
				},
			},
		},
	}

	return response, nil
	/**
	    return new Promise((resolve, reject) => {
			const countryData = request.body.queryResult.parameters["geo-country-code"];
			const countryCode = countryData["alpha-2"];
			const countryName = countryData.name;
			https.get('https://corona.lmao.ninja/v2/countries/' + countryCode, (resp) => {
			  console.log("RESPNSE");
			  let data = '';

			  // A chunk of data has been recieved.
			  resp.on('data', (chunk) => {
				data += chunk;
			  });

			  // The whole response has been received. Print out the result.
			  resp.on('end', () => {
					let parsedData = JSON.parse(data);
					agent.add(parsedData.deaths + ` people have died in ` + countryName + ` from coronavirus.` );
				  return resolve();
			  });

			}).on("error", (err) => {
			  console.log("Error: " + err.message);
			});
		  });
	**/
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
