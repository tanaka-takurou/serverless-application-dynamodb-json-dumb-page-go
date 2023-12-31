package main

import (
	"os"
	"log"
	"context"
	"strings"
	"net/http"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dynamodbClient *dynamodb.Client

func HandleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	var err error
	var bodyString string
	if &request.RequestContext != nil && &request.RequestContext.HTTP != nil && len (request.RequestContext.HTTP.SourceIP) > 0 {
		log.Println(request.RequestContext.HTTP.SourceIP)
	}
	if strings.HasPrefix(request.PathParameters["proxy"], "table") {
		tableName  := request.PathParameters["proxy"][6:]
		log.Println(tableName)
		tableContents, err := getTableContents(ctx, tableName)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "",
			}, err
		}
		jsonBytes, err := json.Marshal(tableContents)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       "",
			}, err
		}
		bodyString = string(jsonBytes)
		return events.APIGatewayProxyResponse{
			StatusCode:      http.StatusOK,
			IsBase64Encoded: false,
			Body:            bodyString,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
	tableNameList, err := getTableNameList(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "",
		}, err
	}
	bodyString = "<html><head><title>Serverless Application Dynamodb Json Dumb</title></head><body><h2>Table List</h2><div><ul>"
	for _, i := range tableNameList {
		bodyString = bodyString + "<li><a href='./table/" + i + "'>" + i + "</a></li>"
	}
	bodyString = bodyString + "</div></ul></body></html>"
	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		IsBase64Encoded: false,
		Body:            bodyString,
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
	}, nil
}

func getTableNameList(ctx context.Context)([]string, error) {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	result, err := dynamodbClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}
	return result.TableNames, nil
}

func getTableContents(ctx context.Context, tableName string)(interface{}, error) {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.NewFromConfig(getConfig(ctx))
	}
	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := dynamodbClient.Scan(ctx, params)
	if err != nil {
		return nil, err
	}
	var tableContents []interface{}
	for _, i := range result.Items {
		tableContents = append(tableContents, i)
	}
	return tableContents, nil
}

func getConfig(ctx context.Context) aws.Config {
	var err error
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv("REGION")))
	if err != nil {
		log.Print(err)
	}
	return cfg
}

func main() {
	lambda.Start(HandleRequest)
}
