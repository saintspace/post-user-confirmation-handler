package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/google/uuid"
)

func main() {
	lambda.Start(handlePostConfirmation)
}

type Event struct {
	EventName     string `json:"eventType"`
	CorrelationId string `json:"correlationId"`
	EventDetails  string `json:"eventDetails"`
}

type AccountConfirmationTaskEvent struct {
	UserName     string `json:"userName"`
	EmailAddress string `json:"emailAddress"`
}

func handlePostConfirmation(event events.CognitoEventUserPoolsPostConfirmation) (events.CognitoEventUserPoolsPostConfirmation, error) {
	userName := event.UserName
	userEmail := event.Request.UserAttributes["email"]
	correlationId := uuid.New().String()
	accountContirmationTask := AccountConfirmationTaskEvent{
		UserName:     userName,
		EmailAddress: userEmail,
	}
	accountConfirmationTaskBytes, err := json.Marshal(accountContirmationTask)
	if err != nil {
		fmt.Printf("error marshalling account confirmation task for (user name: %v) (error: %v)\n", userName, err.Error())
	} else {
		event := Event{
			EventName:     "account-confirmation-task",
			CorrelationId: correlationId,
			EventDetails:  string(accountConfirmationTaskBytes),
		}
		eventBytes, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("error marshalling account confirmation task event for (user name: %v) (error: %v)\n", userName, err.Error())
		}
		eventString := string(eventBytes)
		awsSession := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		ssmService := ssm.New(awsSession)
		paramOutput, err := ssmService.GetParameter(&ssm.GetParameterInput{
			Name:           aws.String("worker-tasks-topic-arn"),
			WithDecryption: aws.Bool(false),
		})
		if err != nil {
			fmt.Printf("error getting worker tasks topic arn to send account confirmation task event for (user name: %v) (error: %v)\n", userName, err.Error())
		} else {
			topicArn := paramOutput.Parameter.Value
			snsService := sns.New(awsSession)
			uniqueMessageId := uuid.New().String()
			_, err := snsService.Publish(&sns.PublishInput{
				Message:                &eventString,
				TopicArn:               topicArn,
				MessageGroupId:         &uniqueMessageId,
				MessageDeduplicationId: &uniqueMessageId,
			})
			if err != nil {
				fmt.Printf("error publishing account confirmation task event for (user name: %v) (error: %v)\n", userName, err.Error())
			} else {
				fmt.Printf("successfully published account confirmation task event for (user name: %v) (correlationId: %v)\n", userName, correlationId)
			}
		}
	}

	return event, nil
}
