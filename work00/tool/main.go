package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ProtocolMapperConfig struct {
	FullPath           string `json:"full.path"`
	IDTokenClaim       string `json:"id.token.claim"`
	AccessTokenClaim   string `json:"access.token.claim"`
	ClaimName          string `json:"claim.name"`
	UserinfoTokenClaim string `json:"userinfo.token.claim"`
}

func NewProtocolMapperConfig() *ProtocolMapperConfig {
	p := new(ProtocolMapperConfig)
	p.FullPath = "true"
	p.IDTokenClaim = "false"
	p.AccessTokenClaim = "true" // たぶんこれがトークンにグループを追加するという意味
	p.ClaimName = "groups"
	p.UserinfoTokenClaim = "true"
	return p
}

type ProtocolMapper struct {
	Name            string               `json:"name"`
	Protocol        string               `json:"protocol"`
	ProtocolMapper  string               `json:"protocolMapper"`
	ConsentRequired bool                 `json:"consentRequired"`
	Config          ProtocolMapperConfig `json:"config"`
}

func NewProtocolMapper() *ProtocolMapper {
	p := new(ProtocolMapper)
	p.Name = "groups"
	p.Protocol = "openid-connect"
	p.ProtocolMapper = "oidc-group-membership-mapper"
	p.ConsentRequired = false
	p.Config = *NewProtocolMapperConfig()
	return p
}

type Client struct {
	ClientID        string           `json:"clientId"`
	PublicClient    bool             `json:"publicClient"`
	RedirectURIs    []string         `json:"redirectUris"`
	WebOrigins      []string         `json:"webOrigins"`
	ProtocolMappers []ProtocolMapper `json:"protocolMappers"`
	Attributes      struct {
		BackchannnelLogoutURL       string `json:"backchannel.logout.url"`
		BackchannnelSessionRequired string `json:"backchannel.logout.session.required"`
	} `json:"attributes"`
}

type Keycloak struct {
	URL      string `json:"url"`
	Account  string `json:"account"`
	Password string `json:"password"`
}

type Config struct {
	Realm    string   `json:"realm"`
	Keycloak Keycloak `json:"keycloak"`
	Clients  []Client `json:"clients"`
	Roles    []string `json:"roles"`
	Groups   []string `json:"groups"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("使い方：tool 設定ファイル")
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(config)
	token := getAccessToken(config)
	deleteRealm(config, token)
	_, err = createRealm(config, token)
	if err != nil {
		log.Fatal(err)
	}
	for _, client := range config.Clients {
		client.ProtocolMappers = []ProtocolMapper{*NewProtocolMapper()}
		respBody, err := createClient(config.Realm, config.Keycloak, client, token)
		if err != nil {
			log.Println(respBody)
			log.Fatal(err)
		}
		log.Println(respBody)
	}
	for _, role := range config.Roles {
		respBody, err := addRole(config, token, role)
		if err != nil {
			log.Println(respBody)
			log.Fatal(err)
		}
	}
	for _, grp := range config.Groups {
		respBody, err := addGroup(config, token, grp)
		if err != nil {
			log.Println(respBody)
			log.Fatal(err)
		}
	}
}

//
//
//
func getAccessToken(c Config) string {
	values := url.Values{}
	values.Set("username", c.Keycloak.Account)
	values.Add("password", c.Keycloak.Password)
	values.Add("grant_type", "password")
	values.Add("client_id", "admin-cli")
	req, err := http.NewRequest(
		"POST",
		c.Keycloak.URL+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(values.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		log.Fatal(fmt.Errorf("%d", resp.StatusCode))
	}
	type Token struct {
		Token string `json:"access_token"`
	}
	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		log.Fatal(err)
	}
	return token.Token
}

//
//
//
func deleteRealm(c Config, token string) {
	log.Println("deleteRealm()")
	del(c.Keycloak.URL+"/admin/realms/"+c.Realm, token)
}

//
//
//
func createRealm(c Config, token string) (string, error) {
	log.Println("createRealm()")
	type RealmInfo struct {
		Realm   string `json:"realm"`
		Enabled bool   `json:"enabled"`
	}
	return post(c.Keycloak.URL+"/admin/realms/", token, RealmInfo{Realm: c.Realm, Enabled: true})
}

func createClient(realm string, k Keycloak, cl Client, token string) (string, error) {
	return post(k.URL+"/admin/realms/"+realm+"/clients", token, cl)
}

func addRole(c Config, token, role string) (string, error) {
	type RoleInfo struct {
		Name string `json:"name"`
	}
	return post(c.Keycloak.URL+"/admin/realms/"+c.Realm+"/roles", token, RoleInfo{Name: role})
}

func addGroup(c Config, token, group string) (string, error) {
	type GroupInfo struct {
		Name string `json:"name"`
	}
	return post(c.Keycloak.URL+"/admin/realms/"+c.Realm+"/groups", token, GroupInfo{Name: group})
}

//
//
//
func post(path, token string, jsondata any) (string, error) {
	log.Println("post(", path, ",", token, ", ...)")
	buf, err := json.Marshal(jsondata)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		"POST", path, bytes.NewBuffer(buf))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		return string(body), fmt.Errorf("%d", resp.StatusCode)
	}
	return string(body), nil
}

func del(path, token string) (string, error) {
	req, err := http.NewRequest(
		"DELETE", path, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		return string(body), fmt.Errorf("%d", resp.StatusCode)
	}
	return string(body), nil
}
