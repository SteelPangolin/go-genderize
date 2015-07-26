package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"fmt"
)

const (
	// The version is used to form the default user agent string.
	version = "0.1.0"

	defaultServer = "https://api.genderize.io/"
)

// Config for a Genderize client.
type Config struct {
	UserAgent  string
	APIKey     string
	Server     string
	HTTPClient *http.Client
}

// The Genderize client itself.
type Genderize struct {
	userAgent  string
	apiKey     string
	apiURL     *url.URL
	httpClient *http.Client
}

// NewGenderize constructs a Genderize client from a Config.
func NewGenderize(config Config) (*Genderize, error) {
	genderize := &Genderize{
		userAgent:  "GoGenderize/" + version,
		apiKey:     config.APIKey,
		httpClient: http.DefaultClient,
	}

	if config.UserAgent != "" {
		genderize.userAgent = config.UserAgent
	}

	if config.HTTPClient != nil {
		genderize.httpClient = config.HTTPClient
	}

	server := defaultServer
	if config.Server != "" {
		server = config.Server
	}
	apiURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}
	genderize.apiURL = apiURL

	return genderize, nil
}

// A NameQuery is a list of names with optional country and language IDs.
type NameQuery struct {
	Names      []string
	CountryID  string
	LanguageID string
}

type NameResponse struct {
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

type rawErrorResponse struct {
	Error string
}

func (genderize *Genderize) Get(nameQuery NameQuery) ([]NameResponse, error) {
	if len(nameQuery.Names) == 0 {
		return nil, nil
	}

	// Build URL query params from NameQuery.
	params := url.Values{}
	for _, name := range nameQuery.Names {
		params.Add("name[]", name)
	}
	if genderize.apiKey != "" {
		params.Add("api_key", genderize.apiKey)
	}
	if nameQuery.CountryID != "" {
		params.Add("country_id", nameQuery.CountryID)
	}
	if nameQuery.LanguageID != "" {
		params.Add("language_id", nameQuery.LanguageID)
	}
	queryURL := *genderize.apiURL
	queryURL.RawQuery = params.Encode()

	// Make the HTTP request.
	req := &http.Request{
		URL: &queryURL,
		Header: http.Header{
			"User-Agent": {genderize.userAgent},
		},
	}
	resp, err := genderize.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Unpack the response.
	decoder := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	raws := []rawNameResponse{}
	if len(nameQuery.Names) == 1 {
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
	nameResponses := make([]NameResponse, len(raws))
	for i, raw := range raws {
		probability, err := strconv.ParseFloat(raw.Probability, 64)
		if err != nil {
			probability = 0.0
		}

		nameResponses[i] = NameResponse{
			Name:        raw.Name,
			Gender:      raw.Gender,
			Count:       raw.Count,
			Probability: probability,
		}
	}
	return nameResponses, nil
}

func main() {
	genderize, err := NewGenderize(Config{})
	if err != nil {
		panic(err)
	}

	nameResponses, err := genderize.Get(NameQuery{
		Names: []string{"Jason", "Molly", "Thunderhorse"},
	})
	if err != nil {
		panic(err)
	}
	for _, nameResponse := range nameResponses {
		fmt.Printf("%#v\n", nameResponse)
	}

	nameResponses, err = genderize.Get(NameQuery{Names: []string{"T-Bone"}})
	if err != nil {
		panic(err)
	}
	for _, nameResponse := range nameResponses {
		fmt.Printf("%#v\n", nameResponse)
	}
}
