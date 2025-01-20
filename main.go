package main

import (
	"fmt"
)

// SimpleHandler prints the parameters to the console
func SimpleHandler(repo string, issueNumber int, title string, body string) {
	fmt.Printf("Repository: %s\n", repo)
	fmt.Printf("Issue Number: %d\n", issueNumber)
	fmt.Printf("Title: %s\n", title)
	fmt.Printf("Body: %s\n", body)
}
