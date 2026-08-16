.PHONY: setup run test format lint gen-db build zip clean tf-fmt

setup: clean
	mise install
	go mod download
	$(MAKE) gen-db

# DB が無ければ生成する
run:
	@test -f assets/master.db || $(MAKE) gen-db
	go run .

test:
	go test -race ./...

format:
	gofmt -w .

# CI と同じバージョンを使う
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run ./...

gen-db:
	rm -f assets/master.db
	sqlite3 assets/master.db < assets/gen.sql

# modernc は純Goのため CGO 不要でクロスコンパイル可
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap .

# Lambda は /var/task に展開するため、DB は assets/master.db の階層を保つ
zip: build gen-db
	rm -f fn.zip
	zip fn.zip bootstrap assets/master.db

clean:
	rm -f bootstrap fn.zip assets/master.db
	go clean -cache -testcache

tf-fmt:
	terraform -chdir=terraform fmt -recursive
