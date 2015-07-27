package genderize

import (
	"fmt"
)

// Simple interface with minimal configuration.
func ExampleGet() {
	responses, err := Get([]string{"James", "Eva", "Thunderhorse"})
	if err != nil {
		panic(err)
	}
	for _, response := range responses {
		fmt.Printf("%s: %s\n", response.Name, response.Gender)
	}
	// Output:
	// James: male
	// Eva: female
	// Thunderhorse:
}

// Client with custom API key and user agent, query with language and country IDs.
func ExampleClient_Get() {
	client, err := NewClient(Config{
		UserAgent: "GoGenderizeDocs/0.0",
		// Note that you'll need to use your own API key.
		APIKey: "",
	})
	if err != nil {
		panic(err)
	}
	responses, err := client.Get(Query{
		Names:      []string{"Kim"},
		CountryID:  "dk",
		LanguageID: "da",
	})
	if err != nil {
		panic(err)
	}
	for _, response := range responses {
		fmt.Printf("%s: %s\n", response.Name, response.Gender)
	}
	// Output:
	// Kim: male
}
