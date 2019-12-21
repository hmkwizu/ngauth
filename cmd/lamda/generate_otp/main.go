package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func lambdaHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//email := req.QueryStringParameters["email"]

	response := make(map[string]interface{})
	response["code"] = http.StatusCreated
	//response["code"] = http.StatusOK
	response["success"] = true
	//response["email"] = email
	response["queryParams"] = req.QueryStringParameters
	response["reqBody"] = req.Body

	// The APIGatewayProxyResponse.Body field needs to be a string, so
	// we marshal the book record into JSON.
	js, err := json.Marshal(response)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(js),
	}, nil
}
func main() {
	lambda.Start(lambdaHandler)
}

// Add a helper for handling errors. This logs any error to os.Stderr
// and returns a 500 Internal Server Error response that the AWS API
// Gateway understands.
func serverError(err error) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       http.StatusText(http.StatusInternalServerError),
	}, nil
}

// Similarly add a helper for send responses relating to client errors.
func clientError(status int) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       http.StatusText(status),
	}, nil
}
