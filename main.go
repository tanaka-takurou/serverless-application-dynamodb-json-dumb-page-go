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
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/dynamodbattribute"
)

var cfg aws.Config
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
		dynamodbClient = dynamodb.New(cfg)
	}
	result, err := dynamodbClient.ListTablesRequest(&dynamodb.ListTablesInput{}).Send(ctx)
	if err != nil {
		return nil, err
	}
	return result.ListTablesOutput.TableNames, nil
}

func getTableContents(ctx context.Context, tableName string)(interface{}, error) {
	if dynamodbClient == nil {
		dynamodbClient = dynamodb.New(cfg)
	}
	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}
	result, err := dynamodbClient.ScanRequest(params).Send(ctx)
	if err != nil {
		return nil, err
	}
	var tableContents []interface{}
	for _, i := range result.ScanOutput.Items {
		var item interface{}
		err := dynamodbattribute.UnmarshalMap(i, &item)
		if err != nil {
			log.Print(err)
		} else {
			tableContents = append(tableContents, item)
		}
	}
	return tableContents, nil
}

func init() {
	var err error
	cfg, err = external.LoadDefaultAWSConfig()
	cfg.Region = os.Getenv("REGION")
	if err != nil {
		log.Print(err)
	}
}

func main() {
	lambda.Start(HandleRequest)
}
