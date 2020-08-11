package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// ※POSTされてくるJSONデータの内容を構造体で定義する
type Request struct {
	UserID      int    	`json:"user_id,string"`
	UserName    string 	`json:"user_name"`
	Age      	int 	`json:"age,string"`
	SendPush    bool  	`json:"send_push,string"`
	SendEmail   bool   	`json:"send_email,string"`
}

// JSONの形のテキストデータをgolangの構造体に変換するための関数
func ConvertInputDataToStruct(inputs string) (*Request, error) {

	var req Request
	err := json.Unmarshal([]byte(inputs), &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

												// ↓ APIGatewayProxyRequest型で引数にPOSTした内容を受け取り、APIGatewayProxyResponseで変換するのが作法
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// テキストデータをgolang構造体に変換
	req, err := ConvertInputDataToStruct(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body: err.Error(),
			StatusCode: 500,
		}, err
	}

	// 変換後の構造体をログに出力する
	fmt.Println("This is Request struct in aws lambda")
	fmt.Printf("(%%#v) %#v\n", req)

	return events.APIGatewayProxyResponse{
		Body: "Success to convert post data to golang struct.",
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
