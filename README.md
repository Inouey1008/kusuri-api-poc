# kusuri-api-poc

AWS Lambda (provided.al2023 / arm64) 上で Go + modernc.org/sqlite (純Go・cgo不要) が動くかを検証する PoC。特にコールドスタート時間を実測で確認する。

**ダミーデータ 3 件のみを使用する。本番の医薬品マスタ (MEDIS 医薬品HOTコードマスター) は含まない。原本・生成した SQLite はコミットしない。**

## アーキテクチャ

機能別パッケージ構成。`internal/` 直下を機能 (`features/`)・共通パッケージ (`sqlite/` `server/` `logging/` `httpx/` `errorx/` `validation/`) に分け、1 機能 = 1 ディレクトリ、層はファイル名で表す。共通パッケージは種類が増えるまで `internal/` 直下にフラットに置く (外部バッキングサービスが複数になったら `infra/` のような分類を検討する)。

```
main.go                        DB 接続 → server 組み立て → 起動 (Lambda/ローカル分岐)
   ▼ internal/server           依存の組み立て (composition root) + ServeMux + middleware
   │                           New(db) が feature と middleware をまとめて配線する
   ▼ internal/logging          logging.Middleware
   ▼ internal/features/drug    医薬品モジュール
        entity.go               Drug エンティティ (JSON タグなし)
        dto.go                  リクエスト/レスポンス DTO + 変換関数
        routes.go                ルート登録 (Register)
        handler.go              HTTP ⇔ JSON 変換 + バリデーション呼び出し
        service.go              検索ロジック
        repository.go           Repository インターフェース
        repository_sqlite.go    sqliteRepository (database/sql + modernc)
   ▼ internal/sqlite     Connect(ctx, path): 読み取り専用で開いて PingContext で疎通確認
   ▼ internal/httpx            WriteJSON (汎用 JSON 書き込み)
   ▼ internal/errorx           Errorx (Internal, Validation) + WithDetails
   ▼ internal/validation       go-playground/validator/v10 ラッパ (FieldError / Validate)
```

拡張ポイント:
- **エンドポイント追加**: `drug/handler.go` にハンドラを足し、`Register` に 1 行加える。server は無変更。
- **モジュール追加**: `internal/features/` 配下に `drug/` と同じ構成のディレクトリを新設し、`server.New` の `[]Registerer` に 1 行足す。既存モジュールには触らない。
- **middleware 追加**: `internal/` 配下にパッケージを作り、`func Middleware(http.Handler) http.Handler` を公開して `server.New` の `[]Middleware` に 1 行足す。先に定義したものが外側 (先に実行される)。
- **DB 差し替え**: `repository_sqlite.go` を同じ `NewRepository` を持つ実装ファイルに差し替える。具象型は非公開なので service/handler/main のいずれも無変更。
- **ローカル実行**: `AWS_LAMBDA_RUNTIME_API` が未セットのときは `:8080` で HTTP サーバを起動する (`PORT` で変更可)。Lambda アダプタを外す必要はない。

## ディレクトリ構成

```
.
├── main.go
├── Makefile
├── go.mod
├── assets/
│   └── gen.sql                         # DB 生成 SQL (ダミーデータ 3 件)
└── internal/
    ├── features/
    │   └── drug/
    │       ├── entity.go               # Drug エンティティ
    │       ├── dto.go                  # searchRequest / drugResponse / searchResponse
    │       ├── repository.go           # Repository インターフェース
    │       ├── repository_sqlite.go    # sqliteRepository (database/sql 実装)
    │       ├── repository_sqlite_test.go
    │       ├── service.go              # 検索ロジック
    │       ├── service_test.go
    │       ├── routes.go               # ルート登録 (Register)
    │       ├── handler.go              # HTTP ハンドラ
    │       └── handler_test.go
    ├── sqlite/
    │   └── sqlite.go                   # Connect (immutable=1, ro) + ドライバ登録
    ├── server/
    │   └── server.go                   # 依存の組み立て + ServeMux + middleware チェーン
    ├── logging/
    │   └── logging.go                  # メソッド・パス・所要時間のログ middleware
    ├── httpx/
    │   └── httpx.go                    # WriteJSON
    ├── errorx/
    │   └── errorx.go                   # Errorx (Internal, Validation) + WithDetails
    ├── validation/
    │   ├── validation.go               # Validate / FieldError
    │   └── validation_test.go
    └── test/
        ├── e2e/
        │   ├── drug_test.go            # GET /drugs のフルスタック検証
        │   └── server_test.go          # feature 横断 (404 など)
        ├── testdb/                     # 一時 SQLite の生成・接続
        └── testrequest/                # 本番と同じ配線で組んだ API へのリクエスト
```

## SQLite 接続

`sqlite.Connect(ctx, path)` は読み取り専用 (`mode=ro&immutable=1`) で開き、`PingContext` で実接続まで確立してから `*sql.DB` を返す。

- `immutable=1` はデプロイ先の読み取り専用 FS (`/var/task`) への対応に必要。
- `PingContext` による fail-fast で DB 欠損・破損を起動時に早期検出する。
- DB パスは `main.go` の定数 `./assets/master.db` に固定 (同梱前提)。

## バリデーション

`internal/validation` は go-playground/validator/v10 のラッパ。DTO の `validate` タグで入力を検証し、違反があれば 400 を返す。

| フィールド | ルール |
|---|---|
| `q` (検索クエリ) | 任意・最大 100 文字 |

エラーレスポンス:

```json
{
  "code": "VALIDATION_FAILED",
  "error": "入力内容に誤りがあります",
  "details": [{"field": "q", "message": "must be at most 100 characters"}]
}
```

## sqlc 導入 (将来)

現状は `internal/features/drug/repository_sqlite.go` を `database/sql` で手書きしている。`drug.Repository` インターフェースが境界になっているため、`sqlc generate` が出力するコードをその実装と差し替えるだけで service/handler は無変更のまま移行できる。

なお sqlc はスキーマ全体から 1 パッケージを生成するのが標準のため、生成コードは `internal/db/` に集約し、`repository_sqlite.go` はそれを呼んで `Drug` へ詰め替える薄いラッパになる想定。

## セットアップ

### DB 生成

```sh
make gen-db
```

### ローカル実行

```sh
make run    # DB がなければ自動生成してから go run .
```

`http://localhost:8080` でリクエストを受け付ける。`PORT` 環境変数でポートを変更可能。

### テスト

```sh
make test   # go test -race ./...
make vet    # go vet ./...
make lint   # golangci-lint (設定は .golangci.yml)
```

push / PR 時は GitHub Actions (`.github/workflows/ci.yml`) で gofmt・vet・lint・
テスト・Lambda 向けクロスコンパイルが実行される。

テストは一時 SQLite ファイルを `t.TempDir()` に生成するため、`assets/master.db` がなくても実行できる。E2E (`internal/test/e2e`) は `server.New(db)` で本番と同じ配線を組み上げて検証する。

### Lambda デプロイ用パッケージ

```sh
make package    # bootstrap (arm64) + assets/master.db を fn.zip にまとめる
```

## API

定義は [openapi.yaml](openapi.yaml) を正典とする (OpenAPI 3.1)。フロントエンドは
`openapi-typescript` などでこのファイルから型を生成できる。

| メソッド | パス | 概要 |
|---|---|---|
| GET | `/drugs` | 薬品名の部分一致検索。`q` は省略可 (省略時は全件・最大 20 件) |

```
GET /drugs?q=エゼチミブ

{"total": 2, "items": [{"yjCode": "2189018F1043", "name": "エゼチミブ錠10mg「JG」"}]}
```

エラーは全エンドポイント共通で `code` / `error` / `details` を返す。`code` は言語非依存の
識別子なので、クライアントの分岐にはこちらを使う。

```
HTTP 400
{"code": "VALIDATION_FAILED", "error": "入力内容に誤りがあります", "details": [{"field": "q", "message": "must be at most 100 characters"}]}
```

## 検証結果

後日 AWS 認証後に追記。

| 項目 | 結果 |
|---|---|
| ローカル: ビルド可否 | OK (`make vet` / `make build`) |
| ローカル: テスト結果 | PASS (`make test`) |
| AWS: Init Duration | 未実施 |
| AWS: Duration | 未実施 |
