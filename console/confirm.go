package console

import (
	"fmt"
	"strings"
)

func Confirm(prompt string) bool {
	fmt.Printf("%s [y/n]: ", prompt)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

	validResponses := []string{"yes", "yep", "y"}
	for _, vr := range validResponses {
		if strings.EqualFold(response, vr) {
			return true
		}
	}

	return false
}
