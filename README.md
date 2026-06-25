# todo-app-go

##  マイグレーションの実行コマンド
```
 migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" up

 // ロールバック
 migrate -source file://migration -database "postgres://todo:todo@localhost:5432/todo?sslmode=disable" down
```