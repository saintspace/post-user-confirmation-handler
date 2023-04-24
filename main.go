package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handlePostConfirmation)
}

func handlePostConfirmation(event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	fmt.Println("Event:", event)
	userEmail := event.Request.UserAttributes["email"]
	fmt.Println("User email:", userEmail)

	return event, nil
}
