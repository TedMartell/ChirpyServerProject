package main

import "testing"

func TestCleanBody(t *testing.T) {
	cases := []struct {
		body     string
		expected string
	}{
		{
			body:     "This is a kerfuffle opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			body:     "This is a SHARBERT opinion I need to share with the world",
			expected: "This is a **** opinion I need to share with the world",
		},
		{
			body:     "This is a SHARBERT!    opinion I need to share with the world",
			expected: "This is a SHARBERT! opinion I need to share with the world",
		},
		// Add more cases here if needed
	}

	for _, cas := range cases {
		actual := cleanBody(cas.body)
		if actual != cas.expected {
			t.Errorf("\nExpected: %s\n actual: %s",
				cas.expected,
				actual,
			)
		}
	}
}
