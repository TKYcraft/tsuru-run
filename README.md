# tsuru-run

ツール系の Discord BOT

## ルール

- `main`リポジトリに変更を加える際は、PR を作成すること。
- `issue`及び`pull request`には適切な Label・Assign を付与すること。
- `pull request`は最低 2 人以上の approve を持って merge とする。(初期段階では柔軟な対応とする)

## 起動方法(Docker コンテナの起動)

1. `.env`ファイルを`Docker/go`直下に作成し、`DISCORD_TOKEN`と`PREFIX`を指定する。
1. Docker 及び docker compose が使用できる状態で、`run.bat`もしくは`run.sh`を実行することで起動する。

## 操作方法

- tsuru.run
  - 接続
- tsuru.dc
  - 切断
