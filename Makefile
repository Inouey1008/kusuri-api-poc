# Go は mise で管理 (mise install で go1.26.5 が入る)。
# sqlite3 はシステムのものを使用。

.PHONY: test vet gen-db build package clean run

# テスト実行
test:
	go test ./...

# 静的チェック
vet:
	go vet ./...

# 検証用 SQLite を生成 (ダミーデータ3件。本番マスタは使わない)
gen-db:
	sqlite3 assets/master.db < assets/gen.sql

# Lambda 用 arm64 バイナリをビルド (modernc は純Goのため CGO 不要でクロスコンパイル可)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap .

# デプロイパッケージ (bootstrap + DB) を zip 化
package: build gen-db
	zip fn.zip bootstrap assets/master.db

# ローカルで HTTP サーバを起動 (http://localhost:8080)。DB が無ければ生成する。
run:
	@test -f assets/master.db || $(MAKE) gen-db
	go run .

# 生成物を削除
clean:
	rm -f bootstrap fn.zip assets/master.db
