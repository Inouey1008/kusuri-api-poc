# kusuri-api-poc

AWS Lambda (provided.al2023 / arm64) 上で Go + modernc.org/sqlite (純Go・cgo不要) が動くかを検証する PoC。特にコールドスタート時間を実測で確認する。

API 定義は [openapi.yaml](openapi.yaml) を参照。

**ダミーデータ 3 件のみを使用する。本番の医薬品マスタ (MEDIS 医薬品HOTコードマスター) は含まない。原本・生成した SQLite はコミットしない。**

## 開発環境

- Go 1.26.5 (mise で管理)
- sqlite3 (システムのものを使用)

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

```sh
make zip
```

`build` (bootstrap の生成) と `gen-db` を実行し、両者を `fn.zip` にまとめる。

生成した `fn.zip` を Lambda にアップロードする。ハンドラは `bootstrap`、ランタイムは `provided.al2023`、アーキテクチャは `arm64`。

## その他コマンド

```sh
make test      # go test -race ./...
make format    # gofmt -w .
make lint      # golangci-lint (設定は .golangci.yml)
make gen-db    # ローカル DB を作り直す
make build     # bootstrap のみ生成 (クロスコンパイルの確認用)
make clean     # 生成物とビルド/テストキャッシュを削除
```

テストは一時 SQLite を `t.TempDir()` に生成するため、`assets/master.db` がなくても実行できる。push / PR 時は GitHub Actions で gofmt・lint・テスト・クロスコンパイルが走る。
