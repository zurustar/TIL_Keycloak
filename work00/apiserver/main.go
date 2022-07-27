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

const SECRETKEY = "HrkADtB2TuYLS9UrEyeWlbSSXrAAigMP"

const CertsURI = "http://192.168.0.200:8080/realms/demo/protocol/openid-connect/certs"
const UserinfoURI = "http://192.168.0.200:8080/realms/demo/protocol/openid-connect/userinfo"

type CertRespData struct {
	Kid     string   `json:"kid"` // key ID
	Kty     string   `json:"kty"` // key type
	Alg     string   `json:"alg"` // algorithm
	Use     string   `json:"use"` // Public Key Use
	N       string   `json:"n"`
	E       string   `json:"e"`
	X5C     []string `json:"x5c"`
	X5T     string   `json:"x5t"`
	X5TS256 string   `json:"x5t#S256"`
}

type CertResp struct {
	Keys []CertRespData `json:"keys"`
}

func getCirts() (CertResp, error) {
	req, err := http.NewRequest("GET", CertsURI, nil)
	if err != nil {
		log.Println(err)
		return CertResp{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return CertResp{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return CertResp{}, err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(err)
		return CertResp{}, err
	}
	log.Println(string(b))
	var data CertResp
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(err)
		return CertResp{}, err
	}
	return data, nil
}

func checkAuthorizationHeader(c *gin.Context) {
	authorizationHeader := c.Request.Header.Get("Authorization")
	if authorizationHeader != "" {
		ary := strings.Split(authorizationHeader, " ")
		if len(ary) == 2 {
			if ary[0] == "Bearer" {
				// イントロスペクションエンドポイントに投げる、
				// これで必要なクライアントの情報を取得できるならかなり良い

				values := url.Values{}
				values.Add("token", ary[1])
				// このへんの値は設定ファイルに変更する必要あり
				u := "http://192.168.0.200:8080/realms/demo/protocol/openid-connect/token/introspect"
				ClientID := "api_server"
				ClientSecret := "ngCEl3yulauREgzJe1uyBDpmhSX8qa0q"
				req, err := http.NewRequest("POST", u, strings.NewReader(values.Encode()))
				if err != nil {
					log.Println(err)
					c.JSON(500, gin.H{})
					c.Abort()
					return
				}
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				s := base64.StdEncoding.EncodeToString([]byte(ClientID + ":" + ClientSecret))
				req.Header.Set("Authorization", "Basic "+s)
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Println(err)
					c.JSON(500, gin.H{})
					c.Abort()
					return
				}
				defer resp.Body.Close()
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					c.JSON(500, gin.H{})
					c.Abort()
					return
				}
				log.Println(string(b))

				// グループとかロールとか入れてみて、どのデータを取り出したいかを実験する
				type Userinfo struct {
					Username    string `json:"username"`
					Active      bool   `json:"active"`
					RealmAccess struct {
						Roles []string `json:"roles"`
					} `json:"realm_access"`
					ResourceAccess struct {
						Account struct {
							Roles []string `json:"roles"`
						} `json:"account"`
					} `json:"resource_access"`
				}
				var user Userinfo
				err = json.Unmarshal(b, &user)
				if err != nil {
					log.Println(err)
					c.JSON(500, gin.H{})
					c.Abort()
					return
				}
				log.Println(user)
				if user.Active {
					c.Set("username", user.Username)
					c.Set("realm_access", `"`+strings.Join(user.RealmAccess.Roles, `","`)+`"`)
					c.Set("resource_access", `"`+strings.Join(user.ResourceAccess.Account.Roles, `","`)+`"`)
				}
			}
		}
	}
	c.Next()
}

func main() {
	log.SetFlags(log.Lshortfile)
	engine := gin.Default()
	engine.GET("/", checkAuthorizationHeader, func(c *gin.Context) {
		username, exists := c.Get("username")
		if exists {
			c.JSON(200, gin.H{"data": "ログインユーザ名：" + username.(string)})
		} else {
			c.JSON(200, gin.H{"data": "ログインしていないよ"})
		}
	})
	engine.Run(":4000")
}
