# kusuri-api-poc

- 添付文書 Pocket の API の PoC
- Go + Echo で実装し、Lambda へデプロイする

## セットアップ

1. [mise](https://mise.jdx.dev/) をインストールする

2. `.env` を生成

```sh
cp .env.example .env
```

3. ツールの導入・依存の取得・ローカル DB の作成をおこなう

```sh
make setup
```

- Go と Terraform のバージョンは [mise.toml](mise.toml) に固定 (CI も同じ値を使う)

4. ローカルサーバー起動

```sh
make run # http://localhost:8080 で起動
```

- API 仕様書は http://localhost:8080/docs で開ける

## 環境変数

- 一覧は [config.go](internal/config/config.go) を参照

### local 環境
- `.env` に書かれた値を mise が読み込み設定

### dev・prod 環境

- GitHub Secrets に登録した値を、Terraform が Lambda の環境変数として設定

## 開発フロー

### デプロイ

- `release/dev`・`release/prod` へのマージで、対応する環境へのインフラ構築・アプリケーションのデプロイをする CI (GitHubActions) が走る

### インフラ

- Terraform で管理する
- `release/*` ブランチへのマージで生成される

### 初回設定

1. `locals.tf` の `subject_prefix` に、OIDC トークンの `sub` のプレフィックスを設定する。プレフィックスは以下で取得する。

```sh
gh api /repos/<owner>/<repo>/actions/oidc/customization/sub --jq '.sub_claim_prefix'
```

2. tfstate の保存先と IAM ロールを作成する

```sh
make tf-setup ENV=dev
make tf-setup ENV=prod
```

3. GitHub に、シークレットを登録する

リポジトリシークレット (dev・prod で共通)

- `AWS_ACCOUNT_ID`: OIDC で引き受けるロールの ARN 組み立てに使用
- `SLACK_TEAM_ID`: 通知先の Slack ワークスペース

環境シークレット (Environments に `dev`・`prod` を作成し、環境ごとに別の値を登録)

- `DOCS_USER` / `DOCS_PASSWORD`: API 仕様書の Basic 認証
- `SLACK_CHANNEL_ID`: アラートの通知先チャンネル (未設定なら連携を作らない)

※ Slack へ通知する場合、ワークスペースの認可は AWS コンソール ( Amazon Q Developer in chat applications ) での手動作業する必要がある。そこで得た ID を上記のシークレットに登録すること。

## その他役立ちコマンド

```sh
make test      # go test -race ./...
make format    # gofmt -w .
make lint      # golangci-lint (設定は .golangci.yml)
make gen-db    # ローカル DB を作り直す
make build     # bootstrap のみ生成 (クロスコンパイルの確認用)
make zip       # build と gen-db を実行し fn.zip にまとめる
make clean     # 生成物とビルド/テストキャッシュを削除

make tf-fmt                   # terraform fmt -recursive
make tf-validate ENV=dev      # Terraform の構文チェック
make tf-plan ENV=dev          # インフラの差分確認
make tf-setup ENV=dev         # tfstate の保存先と IAM ロールを作成 (環境ごとに一度だけ)
```

- `ENV` は既定値を持たない。環境の取り違えを防ぐため、毎回明示する
