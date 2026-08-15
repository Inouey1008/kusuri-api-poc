# kusuri-api-poc

添付文書 Pocket の API 作成に向けて Lambda (provided.al2023 / arm64) 上で Go + SQLite が動くかを検証する PoC

## 開発環境

- Go 1.26.5 (mise で管理)
- sqlite3 (システムのものを使用)
- Terraform 1.15.8 (`.terraform-version` で指定)

## セットアップ

ツールの導入・依存の取得・ローカル DB の作成をまとめて行う。

```sh
make setup
```

## ローカルサーバー起動

```sh
make run    # DB がなければ生成してから起動
```

`http://localhost:8080` で待ち受ける。`PORT` 環境変数で変更可能。

```sh
curl 'http://localhost:8080/drugs?q=エゼチミブ'
```

## インフラ

Terraform で管理する。

### 初回設定

tfstate の保存先を先に作る必要があるため、初回のみ `initial_setup` を apply する。

```sh
terraform -chdir=terraform/initial_setup/env/dev init
terraform -chdir=terraform/initial_setup/env/dev apply
```

### インフラの変更

メモリサイズやタイムアウトなどを変えるときに使う。

```sh
make tf-init ENV=dev     # 初回のみ
make tf-plan ENV=dev
make tf-apply ENV=dev
```

エンドポイントは `function_url` として出力される。コールドスタートを計測する際は `terraform/env/dev/locals.tf` の `memory_size` を変えて apply し直す。Init Duration は CloudWatch Logs で確認する。

## デプロイ

コードだけを更新する。Terraform はコードの差分を見ないため (`ignore_changes`)、インフラに変更がなければ apply は不要。

```sh
make deploy ENV=dev
```

`make zip` で `fn.zip` を作り、`aws lambda update-function-code` で反映する。

## その他コマンド

```sh
make test      # go test -race ./...
make format    # gofmt -w .
make lint      # golangci-lint (設定は .golangci.yml)
make gen-db    # ローカル DB を作り直す
make build     # bootstrap のみ生成 (クロスコンパイルの確認用)
make zip       # build と gen-db を実行し fn.zip にまとめる
make clean     # 生成物とビルド/テストキャッシュを削除
make tf-fmt    # terraform fmt -recursive
```
