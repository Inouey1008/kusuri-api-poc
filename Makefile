# Go は mise で管理 (mise install で go1.26.5 が入る)。
# sqlite3 はシステムのものを使用。

.PHONY: test vet lint gen-db build package clean run

# テスト実行
test:
	go test -race ./...

# 静的チェック
vet:
	go vet ./...

# golangci-lint (設定は .golangci.yml)
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run ./...

# 検証用 SQLite を生成 (ダミーデータ3件。本番マスタは使わない)
# 再実行できるよう作り直す
gen-db:
	rm -f assets/master.db
	sqlite3 assets/master.db < assets/gen.sql

# Lambda 用 arm64 バイナリをビルド (modernc は純Goのため CGO 不要でクロスコンパイル可)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap .

# デプロイパッケージ (bootstrap + DB) を zip 化
# Lambda は /var/task に展開するため、DB は assets/master.db の階層を保つ
package: build gen-db
	rm -f fn.zip
	zip fn.zip bootstrap assets/master.db
	@unzip -l fn.zip

# ローカルで HTTP サーバを起動 (http://localhost:8080)。DB が無ければ生成する。
run:
	@test -f assets/master.db || $(MAKE) gen-db
	go run .

# 生成物を削除
clean:
	rm -f bootstrap fn.zip assets/master.db
