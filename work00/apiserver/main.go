package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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
	log.Println("checkAuthorizationHeader(c)")
	authorizationHeader := c.Request.Header.Get("Authorization")
	if authorizationHeader != "" {
		ary := strings.Split(authorizationHeader, " ")
		if len(ary) == 2 {
			if ary[0] == "Bearer" {
				token, err := jwt.Parse(ary[1], func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(SECRETKEY), nil // ここで公開鍵を返すようにすればいいみたい。で、公開鍵ってどこにあるの？
				})
				if err == nil {
					log.Println("うまくいった！")
					log.Println(token)
				} else {
					log.Println("うまくいかなかった")
					log.Println(err)
				}
			}
		}
	}
	c.Set("username", "")
	c.Next()
}

func main() {
	log.SetFlags(log.Lshortfile)
	data, err := getCirts() // keycloakから鍵情報を取得する、たぶんaccessTokenのデコードにつかうのではないか
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(data)
	engine := gin.Default()
	engine.GET("/", checkAuthorizationHeader, func(c *gin.Context) {
		c.JSON(200, gin.H{"data": "ログインしている場合としていない場合で値を分ける予定だ"})
	})
	engine.Run(":4000")
}
