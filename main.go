package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/router"
	"github.com/inouey1008/kusuri-api-poc/internal/sqlite"
)

// 同梱 DB の固定パス。
const dbPath = "./assets/master.db"

func main() {
	// DB を開いて接続確立。
	db, err := sqlite.Connect(context.Background(), dbPath)
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}

	repo := drug.NewRepository(db)
	svc := drug.NewService(repo)
	h := drug.NewHandler(svc)
	r := router.New(h)

	// Lambda 環境かローカルかを環境変数で判定
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		// Function URLs は API Gateway v2 ペイロード形式のため NewV2 を使う。
		lambda.Start(httpadapter.NewV2(r).ProxyWithContext)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("local server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
