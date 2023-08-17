# SvelteKit を使ってプロジェクトを作る

プロジェクトを作る。app は任意の名前で、そのままフォルダの名前になるので適切な名称にすること。

https://kit.svelte.dev/ に書いてあるように、以下のコマンドを実行する。

```
npm create svelte@latest app
```

上記コマンドを実行すると CLI ベースで色々聞かれるが、注意点としては、TypeScript を使う、Skelton を使う（Demo アプリはビルドでエラーになる！）、くらいか。
続けて以下を実行。静的サイトとして運用する前提。

```
cd app
npm install
npm i -D @sveltejs/adapter-static
```

svelte.config.js を以下のように変更。fallback のところが公式サイトの説明とは違うので注意。

```
import adapter from '@sveltejs/adapter-static';
export default {
    kit: {
        adapter: adapter({
            pages: 'build',
            assets: 'build',
            fallback: 'index.html', // 注意
            precompress: false,
            strict: true
        })
    }
};
```

src/routes/+layout.ts というファイルを作って以下のように記入。公式サイトは拡張子が js になっているけれど、先ほど述べたように TypeScript を選択しているので ts にする。

```
export const prerender = true;
```

ビルドしてみる

```
npm run build
```

続いて実装していくわけだが、様々なログイン状態を各ページで実装していくのは骨の折れる作業なので、keycloak が提供している JavaScript 用ライブラリを使う方が良いだろう。
こちらに説明が書いてあるので、これに従ってやってみる。

https://www.keycloak.org/docs/latest/securing_apps/#_javascript_adapter

まずはインストール。

```
npm install keycloak-js
```
