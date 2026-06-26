# todo-app-go

## テスト用のデータベースについて

テスト用のデータベースはdockerのvolumeにデータがないときに、コンテナを起動することで作成されます。
すでにデータがある場合は一度以下のコマンドでコンテナを削除して、再起動してください。

```sh
docker compose down -v

docker compose up -d
```

##  マイグレーションの実行コマンド

```sh
 migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" up

 // ロールバック
 migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" down
```