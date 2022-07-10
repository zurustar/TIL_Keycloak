# TIL_Keycloak

[https://www.keycloak.org/](Keycloak) の基本的な使い方を把握するために試行錯誤している記録です。

規模の大きいシステムでユーザー登録等を GUI で実施するのは現実的でないように思えたので、極力 API で操作しています。

## Keycloak の起動

[https://www.keycloak.org/](Keycloak) を起動する。[https://www.keycloak.org/getting-started/getting-started-docker](こちら)に従って [https://www.docker.com/](Docker) でやってみる。

ちなみに [https://www.keycloak.org/](Keycloak) のバージョン 17 から起動方法が変わっているので注意。

```
docker run -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:18.0.2 start-dev
```

これを [https://docs.docker.com/compose/](Docker Compose) でやるなら docker-compose.yml はこうなる

```
services:
  keycloak:
    image: quay.io/keycloak/keycloak:18.0.2
    environment:
      - KEYCLOAK_ADMIN=admin
      - KEYCLOAK_ADMIN_PASSWORD=admin
    ports:
      - "8080:8080"
    entrypoint:
      - /opt/keycloak/bin/kc.sh
      - start-dev
```

これで、ブラウザで http://localhost:8080/にアクセスすると [https://www.keycloak.org/](Keycloak) の Web インタフェースにアクセスすることができる。上記コマンドをよく見るとわかるように、管理者のアカウントとパスワードは両方とも admin になっている。

## リバースプロキシの起動

SSO を実現する方法のひとつに、Web アプリの前段に [https://openid.net/connect/](OIDC) に対応したリバースプロキシを設置して、認証回りの処理は全てこいつにやらせるという方法がある。

[https://httpd.apache.org/](Apache)では [https://github.com/zmartzone/mod_auth_openidc](mod_auth_opendic) というモジュールがあるので、これを使ってみることにする。

## 実験用クライアントの環境準備

手元のマシンに [https://www.python.org/](python) が入っていたので [https://www.python.org/](python) でやってみる。
一応仮想環境をつくってから実施する。

```
python -m venv .venv
```

有効化するコマンドは OS によって異なる。以下は Windows の場合。

```
.\.venv\Scripts\activate.bat
```

これを実施してから、このリポジトリに含まれる [https://github.com/zurustar/TIL_Keycloak/blob/main/tool/client.py](client.py) を実行してください。

あとはコードを見てください。レルム一覧取得、レルム削除、レルム作成、レルムロール作成、レルムロール情報取得、グループ作成、ユーザ作成、ユーザ情報取得、ユーザのグループへの追加を実行しています。
ユーザ作成時に直接ロールに追加できるのではないだろうか？と疑っていますが今のところよくわからず、引き続き調査中です。

## リバースプロキシの設定
