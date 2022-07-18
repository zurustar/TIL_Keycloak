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

起動が完了したっぽい雰囲気になったら、起動したのと同一のマシンにてブラウザを用いて [http://localhost:8080/](http://localhost:8080/) にアクセスすると [Keycloak](https://www.keycloak.org/) の Web インタフェースを利用できる。上記コマンドをよく見るとわかるように、管理者のアカウントとパスワードは起動時に環境変数で与えていて、両方とも admin になっている。

ちなみに [Keycloak](https://www.keycloak.org/) のバージョン 17 から起動方法が変わっているので注意。これを書いている時点で唯一と思われる日本語書かれた Keycloak の書籍はバージョン 15 を使用しているので、現行バージョンとは違う起動の仕方が説明されている。

## 実験用データの環境準備

[個人情報テストデータジェネレーター](https://testdata.userlocal.jp/)を利用させていただき個人情報のダミーデータを生成して Keycloak の API を使ってユーザ登録などを実施するツールを作った。

※ちなみに同姓同名のレコードが含まれる場合がある。ユニークなキーであると想定してデータ登録すると失敗するので注意。

go の環境を準備してから(Windows や mac ならインストーラが提供されているのでかんたんに準備できるはず)、このリポジトリに含まれる [tool/main.go](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/main.go) を実行する（コマンドプロンプトで tool 配下に移動して go run ./main.go、あるいは go build して生成される実行ファイルを用いても良い）。このツールはレルム一覧取得、レルム削除、レルム作成、レルムロール作成、レルムロール情報取得、グループ作成、ユーザ作成、ユーザ情報取得、ユーザのグループへの追加、ユーザへのロールの追加、クライアントの登録を実行している。詳細は[ソースコード](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/main.go)を参照すること。

## リバースプロキシの起動

SSO を実現する方法のひとつに、Web アプリの前段に [OIDC](https://openid.net/connect/) に対応したリバースプロキシを設置して、認証回りの処理は全てこいつにやらせるという方法がある。各アプリに OIDC を実装する必要がないのと、Keycloak が公式に提供している OpenID Connect 用クライアントアダプタがフロントエンドの JavaScript 向けのもの以外は廃止されている([ダウンロードページ](https://www.keycloak.org/downloads)にいくと JavaScript 以外 DEPRECATED となっている)ので、アプリで実装するのは避けたほうがいいのではないかと個人的に感じている。

[Apache](https://httpd.apache.org/)では [mod_auth_opendic](https://github.com/zmartzone/mod_auth_openidc) というモジュールがあるので、これを使ってみることにする。

このリバースプロキシに認証周りの処理を実行してほしいので、クライアントとして [Keycloak](https://www.keycloak.org/)に登録する…というのは実はこの前に実行している[ツール](https://github.com/zurustar/TIL_Keycloak/blob/main/tool/main.go)の中で実施済み。手動で実施する場合は、[Keycloak](https://www.keycloak.org/) の [管理コンソール](http://localhost:8080/) に管理者でログインして左メニューの Clients をクリックして[表示される画面](http://localhost:8080/admin/master/console/#/realms/jikken/clients)で適宜入力すればよい。

まず Apache に mod-auth-openidc を導入する[Dockerfile](https://github.com/zurustar/TIL_Keycloak/blob/main/reverse_proxy/Dockerfile)を[こちら](https://qiita.com/Esfahan/items/e44c9b866cb037034541)から頂いた。

続いて[Apache の設定ファイル](https://github.com/zurustar/TIL_Keycloak/blob/main/reverse_proxy/conf/openidc.conf)も[こちら](https://qiita.com/Esfahan/items/e44c9b866cb037034541)から頂いた。

※引き続き[こちら](https://qiita.com/Esfahan/items/e44c9b866cb037034541)を勉強させていただいている中。

# ダミー API サーバ

[Wikipedia](https://ja.wikipedia.org/wiki)の[データ](https://dumps.wikimedia.org/jawiki)でも提供しようかと考えているが、データが巨大なのでちがう良いものがあったらそちらを使う。

※ Python で書いていたのだが m1 mac で psycopg2 が上手いこと使えず、やむをえず go に変更。あわせてデータ投入ツールも python から go に変更した。
