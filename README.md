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

## デプロイ

`release/dev` `release/prod` へのマージで、対象環境へ自動的に反映される。

インフラの差分は `release/*` への PR に plan の結果がコメントされるので、マージ前にそこで確認する。

## インフラ

Terraform で管理する。整形は `make tf-fmt`、構文チェック (`terraform validate`) は PR 時に CI が実行する。

### 差分の確認

```sh
terraform -chdir=terraform/env/dev init    # 初回のみ
terraform -chdir=terraform/env/dev plan
```

### 初回設定

`locals.tf` の `subject_prefix` に、OIDC トークンの `sub` のプレフィックスを設定する。GitHub は名前変更に耐えるよう数値 ID を含めるため、以下で取得する。

```sh
gh api /repos/<owner>/<repo>/actions/oidc/customization/sub --jq '.sub_claim_prefix'
```

初回のみ `initial_setup` を apply して、tfstate の保存先と IAM ロールを作成する

```sh
terraform -chdir=terraform/initial_setup/env/dev init
terraform -chdir=terraform/initial_setup/env/dev apply

terraform -chdir=terraform/initial_setup/env/prod init
terraform -chdir=terraform/initial_setup/env/prod apply
```

続けて GitHub 側を設定する。

- リポジトリシークレット `AWS_ACCOUNT_ID` に AWS アカウント ID を登録する
- `release/dev` `release/prod` ブランチを作成する

### IAM ロール

plan と apply で権限を分けている。apply ロールは対象ブランチへの push でのみ引き受けられる。

| ロール | 権限 | 引き受け条件 |
|---|---|---|
| `<prefix>-terraform-plan` | ReadOnly + tfstate 読み書き | PR |
| `<prefix>-terraform-apply` | Lambda / IAM / Logs + tfstate 読み書き | `release/<env>` への push |

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
