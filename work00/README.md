# このフォルダについて

このフォルダは、keycloak まわりがいまいち理解できている感がないので、Keycloak 以外にアダプタの類を使用せずに自力で実装することでちゃんと理解しようとしているものです。

# 構成

全部 127.0.0.1 で起動する。

使用するポートは以下の通り。

Keycloak - 8080
アプリ（Client もしくは RP) - 5000
リソースサーバ(API サーバ) - 4000

レルム名 - demo
アプリのクライアント ID - kakeibo
リソースサーバのクライアント ID - api_server

# つかいかた

keycloak を起動する。

```
docker-compose up
```

- tool フォルダ配下で go run ./main.go ./config.json を実行する

- keycloak の画面にアクセスして、ユーザを作成し、パスワードの設定と、適当なロールの追加と適当なグループの追加を実施する

- client/config.json を修正する。ClientSecret は keycloak の画面で client の kakeibo の secret をコピー。

- apiserver/main.go を修正する。ClientSecret は keycloak の画面で client の api_server の secret をコピー。★ ここはいずれ config に修正する予定。

- apiserver フォルダ配下で go run ./main.go を実行する

- client フォルダ配下で go run ./main.go ./config.json を実行する

- ブラウザで client の URL にアクセスしてログインなどをやってみる。

- できていないこと：api サーバでトークンから group の情報を取得することができていない。keycloak でトークンにグループ情報を含めるように設定できるのではないか？と疑っている
