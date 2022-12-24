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
			b, err := postIntrospectionEndpoint(ary[1])
			if err != nil {
				log.Println(err)
				c.JSON(500, gin.H{})
				c.Abort()
				return
			}
			active, username, realmRoles, resourceRoles, err := parseUserinfo(b)
			if err != nil {
				log.Println(err)
				c.JSON(500, gin.H{})
				c.Abort()
				return
			}
			if active {
				c.Set("username", username)
				c.Set("realm_access", `"`+strings.Join(realmRoles, `","`)+`"`)
				c.Set("resource_access", `"`+strings.Join(resourceRoles, `","`)+`"`)
			}
		}
	}
	c.Next()
}

func postIntrospectionEndpoint(token string) ([]byte, error) {
	values := url.Values{}
	values.Add("token", token)
	// このへんの値は設定ファイルに変更する必要あり
	u := "http://127.0.0.1:8080/realms/demo/protocol/openid-connect/token/introspect"
	ClientID := "api_server"
	req, err := http.NewRequest("POST", u, strings.NewReader(values.Encode()))
	if err != nil {
		log.Println(err)
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s := base64.StdEncoding.EncodeToString([]byte(ClientID + ":" + secret.ClientSecret))
	req.Header.Set("Authorization", "Basic "+s)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return []byte{}, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return []byte{}, err
	}
	return b, nil
}

func parseUserinfo(b []byte) (bool, string, []string, []string, error) {
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
	err := json.Unmarshal(b, &user)
	if err != nil {
		log.Println(err)
		return false, "", []string{}, []string{}, err
	}
	log.Println(user)
	return user.Active, user.Username, user.RealmAccess.Roles, user.ResourceAccess.Account.Roles, nil
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
