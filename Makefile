.PHONY: setup run test format lint gen-db build zip clean tf-fmt guard-env tf-init tf-validate tf-plan tf-setup

guard-env:
	@test -n "$(ENV)" || { echo "ENV を指定してください (例: make tf-plan ENV=dev)"; exit 1; }

setup: clean
	mise install
	go mod download
	$(MAKE) gen-db

run:
	@test -f assets/master.db || $(MAKE) gen-db
	go run .

test:
	go test -race ./...

format:
	gofmt -w .

lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2 run ./...

gen-db:
	rm -f assets/master.db
	sqlite3 assets/master.db < assets/gen.sql

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -trimpath -ldflags="-s -w" -o bootstrap .

# Lambda は /var/task に展開するため、DB は assets/master.db の階層を保つ必要あり
zip: build gen-db
	rm -f fn.zip
	zip fn.zip bootstrap assets/master.db

clean:
	rm -f bootstrap fn.zip assets/master.db
	go clean -cache -testcache

tf-fmt:
	terraform -chdir=terraform fmt -recursive

tf-init: guard-env
	terraform -chdir=terraform/env/$(ENV) init

tf-validate: tf-init
	terraform -chdir=terraform/env/$(ENV) validate

tf-plan: tf-init
	terraform -chdir=terraform/env/$(ENV) plan

tf-setup: guard-env
	terraform -chdir=terraform/initial_setup/env/$(ENV) init
	terraform -chdir=terraform/initial_setup/env/$(ENV) apply
