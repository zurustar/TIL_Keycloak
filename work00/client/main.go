package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// **************************************************************************
//
// 環境やkeycloakの設定に依存する値たち
//

// keycloakのURL
const KeycloakURL = "http://192.168.0.200:8080" 
// keycloakで作成する必要があるrealmの名前
const MyRealm = "demo"
// keycloakの上記realmに対して登録したこのアプリを示すID
const ClientID = "kakeibo"
// クライアント登録後にkeycloakの画面からコピーしてくる
const ClientSecret = "rP1bvglQwR3jfF2KP2A7jOyxxQen8gjg"
// 認可エンドポイントからリダイレクトされてくるアドレス
const RedirectURI = "http://192.168.0.115:5000/callback"


// **************************************************************************
//
// トークンを取得する
//   このAPからkeycloakに対して直接POSTする
//
type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	RefreshExpiresIn int `json:"refresh_expires_in"`
	TokenType string `json:"token_type"`
	NotBeforePolicy int `json:"not-before-policy"`
	SessionState string `json:"session_state"`
	Scope string `json:"scope"`
}

func getToken(realm, clientID, clientSecret, code, redirectURI string) (Token, error) {
	tokenEndpoint := KeycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
	values := url.Values{}
	values.Add("grant_type","authorization_code")
	values.Add("code", code)
	values.Add("redirect_uri", redirectURI)
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		log.Println(err)
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic " + s)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return Token{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return Token{}, err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(err)
		return Token{}, err
	}
	var token Token
	err = json.Unmarshal(b, &token)
	if err != nil {
		log.Println(err)
		return Token{}, err
	}
	return token, nil
}


// **************************************************************************
//
//
//
//
//
func main() {
	engine := gin.Default()
	engine.Static("/static", "./static")
	engine.LoadHTMLGlob("templates/*")

    // --------------------------------------------------------------------------
	//
	// ログイン要求を受けたらkeycloakにリダイレクトする
	//
	engine.GET("/login", func(c *gin.Context){
		authEndpoint := KeycloakURL
		authEndpoint += "/realms/" + MyRealm + "/protocol/openid-connect/auth" // 書籍のURLはいまのKeycloakでは動かない、要注意
		authEndpoint += "?response_type=code" // 認可コードフロー
		authEndpoint += "&client_id=" + ClientID // クライアント＝このサーバのID、keycloakに登録しておく必要あり
		authEndpoint += "&redirect_uri=" + url.QueryEscape(RedirectURI) // リダイレクトURI
		c.Redirect(302, authEndpoint)
	})

    // --------------------------------------------------------------------------
	//
	// ログイン成功時にコールバックを受けるところ。認可コードが渡されるのでそれを使ってKeycloakにTokenを貰いにいく
	//
	engine.GET("/callback", func(c *gin.Context){
		log.Println("keycloakに渡していたリダイレクトURL、ここに戻ってくるということは認証がうまくいったのだろう")
		sessionState := c.Query("session_state")
		code := c.Query("code")
		log.Println("Keycloakのセッション管理用文字列 ->", sessionState)
		log.Println("認可コード ->", code)
		// 認可コードを使ってトークンを取りに行く
		token, err := getToken(MyRealm, ClientID, ClientSecret, code, RedirectURI)
		if err != nil {
			c.HTML(500, "error.html",gin.H{})
			return
		}
		// とってきたトークンをAPIサーバにアクセスするときに使う。
		// またトークンの有効期限が近づいていたら先にリフレッシュトークンを使って
		// あらたなトークンを取得する必要があると思われる。
		// さらにログオフのときにはそのトークンの削除とkeycloakへのログオフの通知をする必要あり。
		c.HTML(200, "content.html", gin.H{"sessionState": sessionState, "code": code, "accessToken": token.AccessToken})
	})



    // --------------------------------------------------------------------------
	//
	// トップページ
	//
	engine.GET("/", func(c *gin.Context){
		c.HTML(200, "index.html", gin.H{})
	})

	engine.Run(":5000")
}