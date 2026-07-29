# todo-app-go

Go (Echo v5) + PostgreSQL で構築したTodoアプリのバックエンドAPIです。
JWT + Cookieによる認証、CSRF対策、レイヤードアーキテクチャ（controller / usecase / repository / domain）を意識して実装しています。

## 技術スタック

- 言語: Go 1.26
- Webフレームワーク: [Echo v5](https://github.com/labstack/echo)
- ORM: [GORM](https://gorm.io/)（PostgreSQLドライバ）
- 認証: [golang-jwt](https://github.com/golang-jwt/jwt) / [echo-jwt](https://github.com/labstack/echo-jwt)（JWTをCookieで管理）
- バリデーション: [ozzo-validation](https://github.com/go-ozzo/ozzo-validation)
- マイグレーション: [golang-migrate](https://github.com/golang-migrate/migrate)
- 環境変数管理: [godotenv](https://github.com/joho/godotenv)
- Lint: [golangci-lint](https://github.com/golangci/golangci-lint)
- DB: PostgreSQL（Docker Compose）
- CI: GitHub Actions

## ディレクトリ構成

```text
.
├── controller/   HTTPリクエストのハンドリング
├── usecase/      ビジネスロジック・バリデーション
├── repository/   DBアクセス（GORM）
├── domain/       エンティティ・リポジトリインターフェース
├── router/       ルーティング定義
├── middleware/   アクセスログ・リクエストID付与などの独自ミドルウェア
├── db/           DB接続
├── migration/    DBマイグレーションファイル
└── schema/       OpenAPIスキーマ（サブモジュール: todo-app-schema）
```

`domain`にエンティティとリポジトリインターフェースを定義し、`controller` → `usecase` → `domain`、`repository` → `domain`という依存方向で、`repository`が`domain`のインターフェースを実装する構成（依存性逆転の原則）を採用しています。

`schema`はAPI仕様（OpenAPI）を管理する[todo-app-schema](https://github.com/matsukawakohei/todo-app-schema)リポジトリをGitサブモジュールとして参照しています。

## 主な機能

- ユーザー登録・ログイン・ログアウト・ログイン中判定（HttpOnly属性を付与したCookieにJWTを格納）
- CSRF対策（Cookie/Header方式）
- ラベルのCRUD（作成・一覧取得・削除）
- タスクのCRUD（作成・一覧取得・更新・削除）
- アクセスログ出力（リクエストID付与、構造化ログ）

## APIエンドポイント

| Method | Path | 認証 | 説明 |
| --- | --- | --- | --- |
| POST | /api/v1/auth/signup | 不要 | ユーザー登録 |
| POST | /api/v1/auth/login | 不要 | ログイン |
| POST | /api/v1/auth/logout | 不要 | ログアウト |
| POST | /api/v1/auth/me | 必要 | ログイン中判定（ログインユーザー情報取得） |
| GET | /api/v1/labels | 必要 | ラベル一覧取得 |
| POST | /api/v1/labels | 必要 | ラベル作成 |
| DELETE | /api/v1/labels/:id | 必要 | ラベル削除 |
| GET | /api/v1/tasks | 必要 | タスク一覧取得 |
| POST | /api/v1/tasks | 必要 | タスク作成 |
| PUT | /api/v1/tasks/:id | 必要 | タスク更新 |
| DELETE | /api/v1/tasks/:id | 必要 | タスク削除 |

## 環境構築

### 1. サブモジュールの取得

`schema`（OpenAPIスキーマ）はサブモジュールのため、以下のコマンドで取得します。

```sh
git submodule update --init --recursive
```

クローン時にまとめて取得する場合は以下を使用してください。

```sh
git clone --recurse-submodules https://github.com/matsukawakohei/todo-app-go.git
cd todo-app-go
```

### 2. 環境変数の設定

`.env` ファイルを作成し、以下を設定します。

```text
DB_SSLMODE=disable
DB_USER=todo
DB_PASSWORD=todo
DB_HOST=localhost
DB_PORT=5432
DB_NAME=todo
FE_URL=http://localhost:5174
SECRET=（32文字以上のランダムな文字列）
```

### 3. DBの起動

```sh
make db-up
```

すでにvolumeにデータがある場合、一度以下のコマンドでコンテナとデータを削除してから起動してください。

```sh
make db-remove
make db-up
```

### 4. マイグレーションの実行

マイグレーションの実行

```sh
make migrate
```

ロールバック

```sh
make migrate-rollback
```

マイグレーションファイルの作成

```sh
make create-migration {ファイル名}
```

### 5. アプリケーションの起動

```sh
make app
```

デフォルトで `:8080` で起動します（`API_PORT` で変更可能）。

## テスト

```sh
make test
```

テスト用DBは `make db-up` 実行時に自動で作成されます。詳細なログが必要な場合は以下を使用してください。

```sh
make test-detail
```

## Lint

```sh
make lint
```

## CI

GitHub Actionsでプルリクエスト時に以下を自動実行しています（[.github/workflows/pr.yaml](.github/workflows/pr.yaml)）。

- `make test` によるテスト実行
- `golangci-lint` による静的解析
