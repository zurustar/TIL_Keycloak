# TIL_Keycloak

Keycloak の基本的な使い方を把握するために試行錯誤している記録です。

規模の大きいシステムでユーザー登録等を GUI で実施するのは現実的でないように思えたので、極力 API で操作しています。

## Keycloak の起動

Keycloak を起動する。[https://www.keycloak.org/getting-started/getting-started-docker](こちら)に従って Docker でやってみる。

ちなみに keycloak のバージョン 17 から起動方法が変わっているので注意。

```
docker run -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:18.0.2 start-dev
```

これで、ブラウザで http://localhost:8080/にアクセスすると keycloak の Web インタフェースにアクセスすることができる。上記コマンドをよく見るとわかるように、管理者のアカウントとパスワードは両方とも admin になっている。

## 実験用クライアントの環境準備

手元のマシンに Python が入っていたので Python でやってみる。
一応仮想環境をつくってから実施する。

```
python -m venv .venv
```

有効化するコマンドは OS によって異なる。以下は Windows の場合。

```
.\.venv\Scripts\activate.bat
```

これを実施してから、このリポジトリに含まれる client.py を実行してください。

あとはコードを見てください。レルム一覧取得、レルム削除、レルム作成、レルムロール作成、レルムロール情報取得、グループ作成、ユーザ作成、ユーザ情報取得、ユーザのグループへの追加を実行しています。
ユーザ作成時に直接ロールに追加できるのではないだろうか？と疑っていますが今のところよくわからず、引き続き調査中です。

## リバースプロキシの設定
