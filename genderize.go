// Package genderize provides a client for the Genderize.io web service.
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

// Version of this library. Used to form the default user agent string.
const Version = "0.2.0"

const defaultServer = "https://api.genderize.io/"

// API requests are now limited to this many names per request.
// See https://genderize.io/#multipleusage for more.
const batchSize = 10

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

// A ServerError contains a message from the Genderize API server.
type ServerError struct {
	Message    string `json:"error"`
	StatusCode int
	RateLimit  *RateLimit
}

// RateLimit holds info on API quotas from rate limit headers.
// See https://genderize.io/#rate-limiting for details.
type RateLimit struct {
	// The number of names allotted for the current time window.
	Limit int64
	// The number of names left in the current time window.
	Remaining int64
	// Seconds remaining until a new time window opens.
	Reset int64
}

// Error returns the error message.
func (serverError ServerError) Error() string {
	return serverError.Message
}

// Get gender info for names with optional country and language IDs.
func (client *Client) Get(query Query) ([]Response, error) {
	n := len(query.Names)
	if n == 0 {
		return nil, nil
	}

	responses := make([]Response, n)
	responseIdx := 0
	for batchStart := 0; batchStart < n; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > n {
			batchEnd = n
		}
		batchResponses, err := client.getBatch(Query{
			Names:      query.Names[batchStart:batchEnd],
			CountryID:  query.CountryID,
			LanguageID: query.LanguageID,
		})
		if err != nil {
			return nil, err
		}
		for _, batchResponse := range batchResponses {
			responses[responseIdx] = batchResponse
			responseIdx++
		}
	}

	return responses, nil
}

func (client *Client) getBatch(query Query) ([]Response, error) {
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
		apiErr := ServerError{
			StatusCode: resp.StatusCode,
		}

		// Get error message. Best effort.
		decoder.Decode(&apiErr)

		// Parse rate limit headers. Best effort.
		limit, limitErr := strconv.ParseInt(resp.Header.Get("X-Rate-Limit-Limit"), 10, 64)
		remaining, remainingErr := strconv.ParseInt(resp.Header.Get("X-Rate-Limit-Remaining"), 10, 64)
		reset, resetErr := strconv.ParseInt(resp.Header.Get("X-Rate-Reset"), 10, 64)
		if limitErr == nil && remainingErr == nil && resetErr == nil {
			apiErr.RateLimit = &RateLimit{
				Limit:     limit,
				Remaining: remaining,
				Reset:     reset,
			}
		}

		return nil, apiErr
	}

	nameResponses := []Response{}
	err = decoder.Decode(&nameResponses)
	if err != nil {
		return nil, err
	}

	return nameResponses, nil
}

// Get gender info for names, using the default client and country/language IDs.
func Get(names []string) ([]Response, error) {
	nameResponses, err := defaultClient.Get(Query{Names: names})
	return nameResponses, err
}
