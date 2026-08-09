package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/inouey1008/kusuri-api-poc/internal/features/drug"
	"github.com/inouey1008/kusuri-api-poc/internal/infra/sqlite"
	"github.com/inouey1008/kusuri-api-poc/internal/router"
)

func main() {
	// sql.Open は遅延接続のため、フルCPU が割り当てられる INIT フェーズのうちに
	// PingContext でファイルオープンまで完了させ、第1リクエストのレイテンシを下げる。
	db, err := sqlite.Open("./assets/master.db")
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalf("DB ping failed: %v", err)
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
