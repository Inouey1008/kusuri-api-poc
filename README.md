# kusuri-api-poc

AWS Lambda (provided.al2023 / arm64) 上で Go + modernc.org/sqlite (純Go・cgo不要) が動くかを検証する PoC。特にコールドスタート時間を実測で確認する。

API 定義は [openapi.yaml](openapi.yaml) を参照。

**ダミーデータ 3 件のみを使用する。本番の医薬品マスタ (MEDIS 医薬品HOTコードマスター) は含まない。原本・生成した SQLite はコミットしない。**

## 開発環境

- Go 1.26.5 (mise で管理)
- sqlite3 (システムのものを使用)

```sh
mise install
```

## セットアップ

検証用 SQLite (ダミーデータ 3 件) を生成する。

```sh
make gen-db
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
make package    # bootstrap (arm64) + assets/master.db を fn.zip にまとめる
```

生成した `fn.zip` を Lambda にアップロードする。ハンドラは `bootstrap`、ランタイムは `provided.al2023`、アーキテクチャは `arm64`。

## その他コマンド

```sh
make test     # go test -race ./...
make vet      # go vet ./...
make lint     # golangci-lint (設定は .golangci.yml)
make build    # Lambda 用 arm64 バイナリのみビルド
make clean    # 生成物を削除
```

テストは一時 SQLite を `t.TempDir()` に生成するため、`assets/master.db` がなくても実行できる。push / PR 時は GitHub Actions で gofmt・vet・lint・テスト・クロスコンパイルが走る。

## 検証結果

後日 AWS 認証後に追記。

| 項目 | 結果 |
|---|---|
| ローカル: ビルド可否 | OK (`make vet` / `make build`) |
| ローカル: テスト結果 | PASS (`make test`) |
| AWS: Init Duration | 未実施 |
| AWS: Duration | 未実施 |
