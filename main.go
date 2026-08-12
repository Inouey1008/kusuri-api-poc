package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/inouey1008/kusuri-api-poc/internal/server"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

const dbPath = "./assets/master.db"

func main() {
	db, err := sqlite.Connect(context.Background(), dbPath)
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}

	handler := server.New(db)

	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		// Function URLs は API Gateway v2 ペイロード形式のため NewV2 を使う。
		lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("local server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
