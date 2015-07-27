package genderize

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// Gender string constants.
const (
	Male    = "male"
	Female  = "female"
	Unknown = ""
)

const (
	// Version of this library. Used to form the default user agent string.
	Version = "0.1.0"

	defaultServer = "https://api.genderize.io/"
)

// Config for a Genderize client.
type Config struct {
	UserAgent  string
	APIKey     string
	Server     string
	HTTPClient *http.Client
}

// Client for the Genderize API.
type Client struct {
	userAgent  string
	apiKey     string
	apiURL     *url.URL
	httpClient *http.Client
}

// NewClient constructs a Genderize client from a Config.
func NewClient(config Config) (*Client, error) {
	client := &Client{
		userAgent:  "GoGenderize/" + Version,
		apiKey:     config.APIKey,
		httpClient: http.DefaultClient,
	}

	if config.UserAgent != "" {
		client.userAgent = config.UserAgent
	}

	if config.HTTPClient != nil {
		client.httpClient = config.HTTPClient
	}

	server := defaultServer
	if config.Server != "" {
		server = config.Server
	}
	apiURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	client.apiURL = apiURL

	return client, nil
}

var defaultClient *Client

func init() {
	var err error
	defaultClient, err = NewClient(Config{})
	if err != nil {
		panic(err)
	}
}

// A Query is a list of names with optional country and language IDs.
type Query struct {
	Names      []string
	CountryID  string
	LanguageID string
}

// A Response is a name with gender and probability information attached.
type Response struct {
	Name string
	// Gender can be "male", "female", or empty,
	// in which case Probability and Count should be ignored.
	Gender      string
	Probability float64
	Count       int64
}

type rawNameResponse struct {
	Name, Gender, Probability string
	Count                     int64
}

// A ServerError contains a message from the Genderize API server.
type ServerError struct {
	Message string `json:"error"`
}

// Error returns the error message.
func (serverError ServerError) Error() string {
	return serverError.Message
}

// Get gender info for names with optional country and language IDs.
func (client *Client) Get(query Query) ([]Response, error) {
	if len(query.Names) == 0 {
		return nil, nil
	}

	// Build URL query params from Query.
	params := url.Values{}
	for _, name := range query.Names {
		params.Add("name[]", name)
	}
	if client.apiKey != "" {
		params.Add("apikey", client.apiKey)
	}
	if query.CountryID != "" {
		params.Add("country_id", query.CountryID)
	}
	if query.LanguageID != "" {
		params.Add("language_id", query.LanguageID)
	}
	queryURL := *client.apiURL
	queryURL.RawQuery = params.Encode()

	// Make the HTTP request.
	req := &http.Request{
		URL: &queryURL,
		Header: http.Header{
			"User-Agent": {client.userAgent},
		},
	}
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Unpack the response.

	success := 200 <= resp.StatusCode && resp.StatusCode < 300
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()

	if !success {
		apiErr := ServerError{}
		err = decoder.Decode(&apiErr)
		if err != nil {
			return nil, err
		}
		return nil, apiErr
	}

	raws := []rawNameResponse{}
	if len(query.Names) == 1 {
		raw := rawNameResponse{}
		err = decoder.Decode(&raw)
		raws = []rawNameResponse{raw}
	} else {
		err = decoder.Decode(&raws)
	}
	if err != nil {
		return nil, err
	}

	// Convert the response to the exported format.
	nameResponses := make([]Response, len(raws))
	for i, raw := range raws {
		probability, err := strconv.ParseFloat(raw.Probability, 64)
		if err != nil {
			probability = 0.0
		}

		nameResponses[i] = Response{
			Name:        raw.Name,
			Gender:      raw.Gender,
			Count:       raw.Count,
			Probability: probability,
		}
	}
	return nameResponses, nil
}

// Get gender info for names, using the default client and country/language IDs.
func Get(names []string) ([]Response, error) {
	nameResponses, err := defaultClient.Get(Query{Names: names})
	return nameResponses, err
}
