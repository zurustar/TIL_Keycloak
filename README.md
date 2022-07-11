# TIL_Keycloak

[Keycloak](https://www.keycloak.org/) の基本的な使い方を把握するために試行錯誤している記録です。

規模の大きいシステムでユーザー登録等を GUI で実施するのは現実的でないように思えたので、[Keycloak](https://www.keycloak.org/) を操作する際には極力 API を使用しています。

## Keycloak の起動

[こちら](https://www.keycloak.org/getting-started/getting-started-docker)に従って [Docker](https://www.docker.com/) でやってみる。

```
docker run -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:18.0.2 start-dev
```

これを [Docker Compose](https://docs.docker.com/compose/) でやるなら [docker-compose.yml](https://github.com/zurustar/TIL_Keycloak/blob/main/docker-compose.yml) はこうなる

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

起動が完了したっぽい雰囲気になったら、起動したのと同一のマシンにてブラウザを用いて [http://localhost:8080/](http://localhost:8080/) にアクセスすると [Keycloak](https://www.keycloak.org/) の Web インタフェースを利用できる。上記コマンドをよく見るとわかるように、管理者のアカウントとパスワードは起動時に環境変数で与えていて、両方とも admin になっている。私の PC では 1 万ユーザ分のデータを使って 10 秒強で実行が完了する。

ちなみに [Keycloak](https://www.keycloak.org/) のバージョン 17 から起動方法が変わっているので注意。これを書いている時点で唯一と思われる日本語書かれた Keycloak の書籍はバージョン 15 を使用しているので、現行バージョンとは違う起動の仕方が説明されている。

## 実験用データの環境準備

[個人情報テストデータジェネレーター](https://testdata.userlocal.jp/)を利用させていただき個人情報のダミーデータを準備し、手元のマシンに [python](https://www.python.org/) が入っていたので [python](https://www.python.org/) で Keycloak の API を使ってユーザ登録などを実施するツールを作った。

一応 python を実行するための仮想環境をつくってから実施する。

```
python -m venv .venv
```

有効化するコマンドは OS によって異なる。 [Windows](https://www.microsoft.com/ja-jp/windows/) で PowerShell を使っている場合は以下のような感じになる。

```
.\.venv\Scripts\activate.ps1
```

これを実施してからこのリポジトリに含まれる [client.py](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/client.py) を実行する。pip install requests しないといけないかもしれない。このツールはレルム一覧取得、レルム削除、レルム作成、レルムロール作成、レルムロール情報取得、グループ作成、ユーザ作成、ユーザ情報取得、ユーザのグループへの追加、クライアントの登録を実行している。詳細は[ソースコード](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/client.py)を参照すること。

現状ではユーザ作成時にグループに登録しているが、ユーザ登録後に別の API でグループに登録するように変更する予定。

## リバースプロキシの起動

SSO を実現する方法のひとつに、Web アプリの前段に [OIDC](https://openid.net/connect/) に対応したリバースプロキシを設置して、認証回りの処理は全てこいつにやらせるという方法がある。

[Apache](https://httpd.apache.org/)では [mod_auth_opendic](https://github.com/zmartzone/mod_auth_openidc) というモジュールがあるので、これを使ってみることにする。

このリバースプロキシに認証周りの処理を実行してほしいので、クライアントとして [Keycloak](https://www.keycloak.org/)に登録する…というのは実はこの前に実行している[ツール](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/client.py)の中で実施済み。手動で実施する場合は、[Keycloak](https://www.keycloak.org/) の [管理コンソール](http://localhost:8080/) に管理者でログインして左メニューの Clients をクリックして[表示される画面](http://localhost:8080/admin/master/console/#/realms/jikken/clients)で適宜入力すればよい。

※現在リバプロ用の Apache を起動する Dockerfile 作成で試行錯誤中。わかったらまた追記する予定。[こちら](https://qiita.com/Esfahan/items/e44c9b866cb037034541)を勉強させていただくとなにかわかりそう。
