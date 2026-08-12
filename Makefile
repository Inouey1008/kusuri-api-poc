.PHONY: setup test format lint gen-db build zip clean run

# 環境を作り直す
setup: clean
	mise install
	go mod download
	$(MAKE) gen-db

test:
	go test -race ./...

format:
	gofmt -w .

# CI と同じバージョンを使う
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run ./...

# ローカル DB を作成
gen-db:
	rm -f assets/master.db
	sqlite3 assets/master.db < assets/gen.sql

# bootstrap (実行ファイル) のみ生成する
# modernc は純Goのため CGO 不要でクロスコンパイル可
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap .

# build + DB を fn.zip にまとめ、Lambda にアップロードできる状態にする
# Lambda は /var/task に展開するため、DB は assets/master.db の階層を保つ
zip: build gen-db
	rm -f fn.zip
	zip fn.zip bootstrap assets/master.db
	@unzip -l fn.zip

# ローカルサーバを起動。DB が無ければ生成する
run:
	@test -f assets/master.db || $(MAKE) gen-db
	go run .

clean:
	rm -f bootstrap fn.zip assets/master.db
	go clean -cache -testcache
