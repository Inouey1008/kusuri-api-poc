# kusuri-api-poc

AWS Lambda (provided.al2023 / arm64) 上で Go + modernc.org/sqlite (純Go・cgo不要) が動くかを検証する PoC。特にコールドスタート時間を実測で確認する。

**ダミーデータ 3 件のみを使用する。本番の医薬品マスタ (MEDIS 医薬品HOTコードマスター) は含まない。原本・生成した SQLite はコミットしない。**

## アーキテクチャ

機能別パッケージ構成。`internal/` 直下を機能 (`features/`)・基盤 (`infra/`)・組み立て (`router/`) の 3 つに分け、1 機能 = 1 ディレクトリ、層はファイル名で表す。

```
main.go                        依存の組み立て (composition root)
   ▼ internal/router           ServeMux (Go 1.22 method+path) + ログ middleware
   │                           Registerer 経由で登録するためモジュールを import しない
   ▼ internal/features/drug    医薬品モジュール
        handler.go              HTTP ⇔ JSON 変換 + ルート登録 (Register)
        service.go              検索ロジック
        repository.go           Repository インターフェース
        repository_sqlite.go    sqliteRepository (database/sql + modernc)
        drug.go                 Drug エンティティ (JSON タグなし)
   ▼ internal/infra/sqlite     接続の生成 (immutable=1, ro) とドライバ登録
```

拡張ポイント:
- **エンドポイント追加**: `drug/handler.go` にハンドラを足し、`Register` に 1 行加える。router は無変更。
- **モジュール追加**: `internal/features/` 配下に `drug/` と同じ構成のディレクトリを新設し、`main.go` で `router.New` に渡す。既存モジュールには触らない。
- **DB 差し替え**: `repository_sqlite.go` を同じ `NewRepository` を持つ実装ファイルに差し替える。具象型は非公開なので service/handler/main のいずれも無変更。
- **Lambda 離脱**: `lambda.Start` と `httpadapter` を外して `http.ListenAndServe` に切り替えるだけ。router 以下は流用できる。

## ディレクトリ構成

```
.
├── main.go
├── go.mod
├── assets/
│   └── gen.sql          # DB 生成 SQL (ダミーデータ 3 件)
└── internal/
    ├── features/
    │   └── drug/
    │       ├── drug.go                 # Drug エンティティ
    │       ├── repository.go           # Repository インターフェース
    │       ├── repository_sqlite.go    # sqliteRepository (database/sql 実装)
    │       ├── repository_sqlite_test.go
    │       ├── service.go              # 検索ロジック
    │       ├── service_test.go
    │       ├── handler.go              # HTTP ハンドラ + DTO + Register
    │       └── handler_test.go
    ├── infra/
    │   └── sqlite/
    │       └── sqlite.go               # SQLite 接続の生成 (ドライバ登録もここ)
    └── router/
        └── router.go                   # ServeMux + middleware + Registerer
```

## sqlc 導入 (将来)

現状は `internal/features/drug/repository_sqlite.go` を `database/sql` で手書きしている。`drug.Repository` インターフェースが境界になっているため、`sqlc generate` が出力するコードをその実装と差し替えるだけで service/handler は無変更のまま移行できる。

なお sqlc はスキーマ全体から 1 パッケージを生成するのが標準のため、生成コードは `internal/db/` に集約し、`repository_sqlite.go` はそれを呼んで `Drug` へ詰め替える薄いラッパになる想定。

## セットアップ

### DB 生成

```sh
sqlite3 assets/master.db < assets/gen.sql
```

### テスト

```sh
mise use go@latest
go test ./...
```

`internal/features/drug` のテストは一時 SQLite ファイルを `t.TempDir()` に生成するため、`assets/master.db` がなくても実行できる。

## API

クエリパラメータ `q` で薬品名の部分一致検索。

```
GET <Function URL>/drugs?q=エゼチミブ
```

レスポンス:

```json
{"total": 2, "items": [{"yjCode": "...", "name": "..."}]}
```

YJ コード直接取得:

```
GET <Function URL>/drugs/{yjCode}
```

## 検証結果

後日 AWS 認証後に追記。

| 項目 | 結果 |
|---|---|
| ローカル: ビルド可否 | OK (`go build ./...` / `go vet ./...`) |
| ローカル: テスト結果 | PASS (`go test ./...`) |
| AWS: Init Duration | 未実施 |
| AWS: Duration | 未実施 |
