package main

import (
	"testing"
)

func TestvalidURLWithHttpUrl(t *testing.T) {
	// Arrange
	valueToTest := "http://test.com"

	// Act
	valid := validURL(valueToTest)

	// Assert
	if !valid {
		t.Errorf("URL not valid. %s", valueToTest)
	}
}

func TestvalidURLWithHttpsUrl(t *testing.T) {
	// Arrange
	valueToTest := "https://test.com"

	// Act
	valid := validURL(valueToTest)

	// Assert
	if !valid {
		t.Errorf("URL not valid. %s", valueToTest)
	}
}

func TestvalidURLWithInvalidURL(t *testing.T) {
	// Arrange
	valueToTest := "/test/test.com"

	// Act
	valid := validURL(valueToTest)

	// Assert
	if valid {
		t.Errorf("URL not valid. %s", valueToTest)
	}
}

func TestvalidURLWithEmptyString(t *testing.T) {
	// Arrange
	valueToTest := ""

	// Act
	valid := validURL(valueToTest)

	// Assert
	if valid {
		t.Errorf("URL not valid. %s", valueToTest)
	}
}
