package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// **************************************************************************
//
// 環境やkeycloakの設定に依存する値たち
//

// keycloakのURL
const KeycloakURL = "http://192.168.0.200:8080"
const ClientURL = "http://192.168.0.115:5000"
const BindAddress = ":5000"

const APIServerURL = "http://192.168.0.115:4000"

// keycloakで作成する必要があるrealmの名前
const MyRealm = "demo"

// keycloakの上記realmに対して登録したこのアプリを示すID
const ClientID = "kakeibo"

// クライアント登録後にkeycloakの画面からコピーしてくる
const ClientSecret = "HrkADtB2TuYLS9UrEyeWlbSSXrAAigMP"

// 認可エンドポイントからリダイレクトされてくるアドレス
const RedirectPath = "/callback"
const RedirectURI = ClientURL + RedirectPath

// **************************************************************************
//
// ログイン要求を受けたらkeycloakにリダイレクトする
//
func procLogin(c *gin.Context) {
	authEndpoint := KeycloakURL
	authEndpoint += "/realms/" + MyRealm + "/protocol/openid-connect/auth" // 書籍のURLはいまのKeycloakでは動かない、要注意
	authEndpoint += "?response_type=code"                                  // 認可コードフロー
	authEndpoint += "&client_id=" + ClientID                               // クライアント＝このサーバのID、keycloakに登録しておく必要あり
	authEndpoint += "&redirect_uri=" + url.QueryEscape(RedirectURI)        // リダイレクトURI
	c.Redirect(302, authEndpoint)
}

// **************************************************************************
//
// トークンを取得する
//   このAPからkeycloakに対して直接POSTする
//
type Token struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func getToken(realm, clientID, clientSecret, code, redirectURI string) (Token, error) {
	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)
	values.Add("redirect_uri", redirectURI)
	return _getToken(realm, clientID, clientSecret, values)
}

// --------------------------------------------------------------------------
//
// ログイン成功時にコールバックを受けるところ。認可コードが渡されるのでそれを使ってKeycloakにTokenを貰いにいく
//
func procCallback(c *gin.Context) {
	sessionState := c.Query("session_state")
	code := c.Query("code")
	log.Println("Keycloakのセッション管理用文字列 ->", sessionState)
	log.Println("認可コード ->", code)
	// 認可コードを使ってトークンを取りに行く
	token, err := getToken(MyRealm, ClientID, ClientSecret, code, RedirectURI)
	if err != nil {
		c.HTML(500, "error.html", gin.H{})
		return
	}
	// access-tokenとrefresh-tokenをクッキーに保存する
	c.SetCookie("access-token", token.AccessToken, token.ExpiresIn, "/", ClientURL, false, true)
	c.SetCookie("refresh-token", token.RefreshToken, token.RefreshExpiresIn, "/", ClientURL, false, true)
	c.HTML(200, "content.html", gin.H{"sessionState": sessionState, "code": code, "accessToken": token.AccessToken})
}

// --------------------------------------------------------------------------
//
// トップページ
//
func procTopPage(c *gin.Context) {
	log.Println("procTopPage(c)")
	accessToken, exists := c.Get("access-token")
	if !exists {
		accessToken = ""
	}
	result, err := getDataFromAPIServer(accessToken.(string))
	if err != nil {
		c.HTML(200, "index.html", gin.H{"data": "failed to get api server"})
		return
	}
	c.HTML(200, "index.html", gin.H{"data": result})
}

// **************************************************************************
//
// APIサーバから情報を取得する
//
func getDataFromAPIServer(accessToken string) (string, error) {
	log.Println("getDataFromAPIServer(", accessToken, ")")
	u := APIServerURL + "/"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(err)
		return "", err
	}
	type Data struct {
		Data string `json:"data"`
	}
	var data Data
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return data.Data, nil
}


// **************************************************************************
//
// トークン関連のミドルウェア、
//   access-tokenがあったらそれを使うし、
//   access-tokenがないのにrefresh-tokenがあったらリフレッシュするし、
//   refresh-tokenもなかったらログアウト後なので特に何もしない
//
func checkToken(c *gin.Context) {
	// APIサーバから情報を取得してきて何かする
	// 1. access-tokenはあるのか？
	accessToken, err := c.Cookie("access-token")
	if err == nil {
		c.Set("access-token", accessToken)
		c.Next()
		return
	}
	// 2. refresh-tokenはあるのか
	refreshToken, err := c.Cookie("refresh-token")
	if err != nil {
		// refresh-tokenがなかったのでたぶんログインしていない。
		c.Next()
		return
	}
	// access-tokenがないのにrefresh-tokenがあったので、リフレッシュする
	log.Println("refresh access-token...")
	values := url.Values{}
	values.Add("grant_type", "refresh_token")
	//	values.Add("client_id", ClientID)
	//	values.Add("client_secret", ClientSecret)
	values.Add("refresh_token", refreshToken)
	token, err := _getToken(MyRealm, ClientID, ClientSecret, values)
	if err != nil {
		// トークン取得に失敗した
		c.Next()
		return
	}
	c.SetCookie("access-token", token.AccessToken, token.ExpiresIn, "/", ClientURL, false, true)
	c.SetCookie("refresh-token", token.RefreshToken, token.RefreshExpiresIn, "/", ClientURL, false, true)
	c.Set("access-token", token.AccessToken)
	c.Next()

}

// **************************************************************************
//
//
//
//
//
func main() {
	log.SetFlags(log.Lshortfile)
	engine := gin.Default()

	// cookieの準備
	store := cookie.NewStore([]byte("secret")) // "secret"は外から持ってきた値を設定する処理にすべき
	store.Options(
		sessions.Options{
			Path:     "/",
			MaxAge:   24 * 60 * 60,
			HttpOnly: true,
		},
	)
	// これがないと落ちるという情報を見かけたが、入れているのでおそらくこれではない
	engine.Use(sessions.Sessions("session", store))

	engine.Static("/static", "./static")
	engine.LoadHTMLGlob("templates/*")
	engine.GET("/login", procLogin)
	engine.GET(RedirectPath, procCallback)
	engine.GET("/", checkToken, procTopPage)
	engine.Run(BindAddress)
}

// **************************************************************************
//
//　トークンを取得する
//
func _getToken(realm, clientID, clientSecret string, values url.Values) (Token, error) {
	tokenEndpoint := KeycloakURL + "/realms/" + realm + "/protocol/openid-connect/token"
	req, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		log.Println(err)
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+s)
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
