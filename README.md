# 概要
[Qrunchに掲載したテック記事用公開ソース]()
API Gateway + AWS Lambda(golang)の組み合わせで、POSTをした際投げたJSONデータをLambda内でキャッチして自由にいじる設定方法と実装方法。

<div align="center">
<img src="https://s3.qrunch.io/c02ddf3f574ecb65d39188fcd0446e89.jpg" alt="属性" title="タイトル">
</div>

# 仕様
curlでPOSTした時、USER_ID, USER_NAME, AGE, SEND_PUSH, SEND_EMAILを受け取り、これをgolangの構造体に変換しログに出力する。
なお、serverless-frameworkの使用を前提とする。

# 結論
以下三つのファイルを用意し、Makefileのある階層で`make deploy`を打てば、標題の機能を備えたLambda + API Gatewayが生成される。

1. serverless.yml
1. main.go
1. Makefile

[GitHubのリポジトリはこちら](https://github.com/mu-editech/go-lambda-postdata-control)
### serverless.yml
```serverless.yml
service: go-lambda-post-control
frameworkVersion: '>=1.28.0 <2.0.0'

provider:
  name: aws
  runtime: go1.x
  stage: dev
  region: ap-northeast-1
  profile: your-profile

package:
  exclude:
    - ./**
  include:
    - ./bin/**

functions:
  go-lambda-post-control:
    handler: bin/go-lambda-post-control
    timeout: 300
    events:
      - http:
          path: eventtest
          method: post

```

### Makefile
```Makefile
.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/go-lambda-post-control main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose

```

### main.go(AWS Lambdaのソース)
```main.go
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

```

# ワークショップ
ここからは説明編。
実際に作りながら、結論で開示したソースの内容を説明していく。

### Step1. セットアップ

まずは以下のコマンドを打ったあと、ファイルを移動する。

```
# 初期スクリプト作成
$ sls create --template aws-go --path go-lambda-post-control --name go-lambda-post-control

# 不要なファイルを削除し、このような形になるようにファイルを移動しつつ、ここにないファイルは不要なので削除する。
go-lambda-post-control/
├── Makefile
├── bin
│   └── hello
├── main.go
└── serverless.yml

```

各ファイルを以下のようにデフォルトの記述に少し追記する。

#### serverless.yml
```serverless.yml
service: go-lambda-post-control
frameworkVersion: '>=1.28.0 <2.0.0'

provider:
  name: aws
  runtime: go1.x
  stage: dev
  region: ap-northeast-1
  profile: your-profile # ~/.aws/credentialsに定めている差し先を指定するために追記。defaultの差し先にDeployしたい場合は不要。


package:
  exclude:
    - ./**
  include:
    - ./bin/**

functions:
  go-lambda-post-control: # lambda関数名に変更する。デフォルトだとhello。
    handler: bin/go-lambda-post-control # ここもmakeファイルのdeployした名前に変更。
    timeout: 300
    events:
      - http:
          path: eventtest
          method: post
```

#### Makefile
```Makefile
.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/go-lambda-post-control main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
```

ここまで終わったら、Deployする。

```
$ make deploy

# 以下のような表示が出てくるので、ここのPOST の部分をコピペし、curlでPOSTすると
・
・
・
Service Information
service: go-lambda-post-control
stage: dev
region: ap-northeast-1
stack: go-lambda-post-control-dev
resources: 11
api keys:
  None
endpoints:
  POST - https://your-endpoint # <= ココをコピーする！！！
functions:
  go-lambda-post-control: go-lambda-post-control-dev-go-lambda-post-control
layers:
  None
・
・
# POSTをしてみると、serverless frameworkがデフォで打ち返す文字列が表示される。
$ curl -X POST -H "Content-Type: application/json" -d '{"user_id": "12345", "user_name":"テストユーザー", "age":"25", "send_push":"true", "send_email": "false"}' https://your-endpoint
{"message":"Go Serverless v1.0! Your function executed successfully!"} 
```

### Step2. Lambdaのソース追記
以下のようにLambdaの中身を書き換える。
処理の順番

1. POSTされてきたデータの変換先となる構造体を定義
1. events.APIGatewayProxyRequest型でPOSTされたデータをキャッチ
1. 上で受け取ったデータの中を、構造体に変換
1. ログ出力

※ `Unmarchal`はJSONの形式になっているテキストを構造体へ変換してくれる優れもの。しかし変換先の構造体の要素にintやboolなどのstring以外のものが入っていた場合、エラーで落ちてしまう。ゆえに``json:"user_id,string"``と言った書き方をしている。
公式はこのように発表している：https://golang.org/pkg/encoding/json/#Marshal

#### main.go
```main.go
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
``` 

これでcurlをすれば、`Success to convert post data to golang struct.`の文字が返ってくるはず。
また、CloudWatchLogsを見れば、POSTした内容が構造体として出力されている。
![undefined.jpg](https://s3.qrunch.io/d9f058b0464c9a39a11cb0c10f27a7da.jpg)

# 余談：<font color="OrangeRed">死ぬほどハマりポイント　lambda-integration</font>
丸2日ほど溶かして悔しかったので共有。

きちんとserverless.ymlに明示してdeployしようとしてserverless.ymlにlambda: integrationをつけるとAPIGatewayProxyRequest.body への変換がこける。こちらを外してもserverless frameworkの方ではAPI GatewayとLambdaをくっつけてくれるので、外して実装した。
### serverless.ymlのラストの部分を抜粋
```
functions:
  go-lambda-post-control:
    handler: bin/go-lambda-post-control
    timeout: 300
    events:
      - http:
          path: eventtest
          method: post
          integration: lambda # <= ここ！書いちゃダメ！！
```

なお、これを書いちゃった場合、`lambda.Start(Handler)`までは処理が到達するが、`func Handler`の中にまでは処理が到達せず、以下のエラーを吐いて落ちる。

```
json: cannot unmarshal object into Go struct field APIGatewayProxyRequest.body of type string: UnmarshalTypeError null
```

英語のIssueのログに同様に悩んでいる海外の人がいた。
[json: cannot unmarshal object into Go struct field APIGatewayProxyRequest.body of type string on GET request #68](https://github.com/aws/aws-lambda-go/issues/68)

おしまい。