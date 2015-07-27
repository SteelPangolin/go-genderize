package genderize

import (
	"testing"
)

func TestIntegration(t *testing.T) {
	responses, err := Get([]string{"James", "Eva", "Thunderhorse"})
	expectedGenders := map[string]string{
		"James":        Male,
		"Eva":          Female,
		"Thunderhorse": Unknown,
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, response := range responses {
		expected := expectedGenders[response.Name]
		actual := response.Gender
		if expected != actual {
			t.Errorf("%s: expected %s, got %s",
				response.Name, expected, actual)
		}
	}
}

func TestInvalidAPIKey(t *testing.T) {
	client, err := NewClient(Config{APIKey: "invalid_api_key"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Get(Query{Names: []string{"Peter"}})
	expectedMsg := "Invalid API key"
	if err == nil || err.Error() != expectedMsg {
		t.Errorf("Expected error with %v, got %#v",
			expectedMsg, err)
	}
}
