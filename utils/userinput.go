package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func GetUserConfirmation(text string) (bool, error) {
	// Display the text string to the user
	fmt.Println(text + " (y/N):")

	// Create a new reader from standard input
	reader := bufio.NewReader(os.Stdin)

	// Read the user's input
	input, err := reader.ReadString('\n')
	if err != nil {
		// Return false and the error if there's an error reading input
		return false, err
	}

	// Trim whitespace and convert input to lower case
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	// Check if the user's input is an affirmative response
	if input == "y" || input == "yes" {
		return true, nil
	}

	// Any other input is considered a negative response
	return false, nil
}
