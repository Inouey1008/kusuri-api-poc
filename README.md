# kusuri-api-poc

添付文書 Pocket の API 作成に向けて Lambda (provided.al2023 / arm64) 上で Go + SQLite が動くかを検証する PoC

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
