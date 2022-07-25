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


const KeycloakURL = "http://192.168.0.200:8080"
const MyRealm = "demo"
const ClientID = "kakeibo" // keycloakに登録したこのアプリのID
const ClientSecret = "rP1bvglQwR3jfF2KP2A7jOyxxQen8gjg" // keycloakの画面からコピーしてくる
const RedirectURI = "http://192.168.0.115:5000/callback"

//
// 認可エンドポイント の URLを取得する
//
func getAuthEndpoint(realm, clientID, RedirectURI string) string {
	u := KeycloakURL
	u += "/realms/" + realm + "/protocol/openid-connect/auth" // 書籍のURLはいまのKeycloakでは動かない、要注意
	u += "?response_type=code" // 認可コードフロー
	u += "&client_id=" + clientID // クライアント＝このサーバのID、keycloakに登録しておく必要あり
	u += "&redirect_uri=" + url.QueryEscape(RedirectURI) // リダイレクトURI
	return u
}

//
// トークンエンドポイントのURLを取得する
//
func getTokenEndpoint(realm string) string {
	return KeycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
}

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
	u := getTokenEndpoint(realm)
	values := url.Values{}
	values.Add("grant_type","authorization_code")
	values.Add("code", code)
	values.Add("redirect_uri", redirectURI)
	req, err := http.NewRequest("POST", u, strings.NewReader(values.Encode()))
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

func main() {
	engine := gin.Default()
	engine.Static("/static", "./static")
	engine.LoadHTMLGlob("templates/*")

	//
	//
	// ログイン要求を受けたらkeycloakにリダイレクトする
	//
	//
	engine.GET("/login", func(c *gin.Context){
		u := getAuthEndpoint(MyRealm, ClientID, RedirectURI)
		log.Println(u)
		c.Redirect(302, u)
	})

	//
	//
	// ログイン成功時にコールバックを受けるところ。認可コードが渡されるのでそれを使ってKeycloakにTokenを貰いにいく
	//
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
		log.Println(token)
		c.HTML(200, "content.html", gin.H{"sessionState": sessionState, "code": code, "accessToken": token.AccessToken})
	})



	//
	//
	//
	engine.GET("/", func(c *gin.Context){
		c.HTML(200, "index.html", gin.H{})
	})

	engine.Run(":5000")
}