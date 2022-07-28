package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Secret struct {
	ClientSecret string `json:"APIServerSecret"`
}

var secret Secret

func checkAuthorizationHeader(c *gin.Context) {
	ary := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(ary) == 2 {
		if ary[0] == "Bearer" {
			// イントロスペクションエンドポイントに投げる、
			// これで必要なクライアントの情報を取得できるならかなり良い

			values := url.Values{}
			values.Add("token", ary[1])
			// このへんの値は設定ファイルに変更する必要あり
			u := "http://192.168.0.200:8080/realms/demo/protocol/openid-connect/token/introspect"
			ClientID := "api_server"
			req, err := http.NewRequest("POST", u, strings.NewReader(values.Encode()))
			if err != nil {
				log.Println(err)
				c.JSON(500, gin.H{})
				c.Abort()
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			s := base64.StdEncoding.EncodeToString([]byte(ClientID + ":" + secret.ClientSecret))
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
				Groups []string `json:"groups"`
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
	c.Next()
}

func main() {
	log.SetFlags(log.Lshortfile)

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(data, &secret)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(secret)

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
