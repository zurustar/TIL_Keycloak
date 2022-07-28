package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

type ClientAttributes struct {
	BackchannnelLogoutURL       string `json:"backchannel.logout.url"`
	BackchannnelSessionRequired string `json:"backchannel.logout.session.required"`
}

type Client struct {
	ClientID        string           `json:"clientId"`
	PublicClient    bool             `json:"publicClient"`
	RedirectURIs    []string         `json:"redirectUris"`
	WebOrigins      []string         `json:"webOrigins"`
	ProtocolMappers []ProtocolMapper `json:"protocolMappers"`
	Attributes      ClientAttributes `json:"attributes"`
}

type Keycloak struct {
	URL           string
	AdminAccount  string
	AdminPassword string
}

type Config struct {
	Realm    string
	Keycloak Keycloak
	Clients  []Client `json:"clients"`
	Roles    []string
	Groups   []string
}

func NewConfig() *Config {
	p := new(Config)
	p.Realm = "demo"
	p.Keycloak = Keycloak{
		URL:           "http://127.0.0.1:8080",
		AdminAccount:  "admin",
		AdminPassword: "admin",
	}
	p.Clients = []Client{
		{
			ClientID:     "kakeibo",
			PublicClient: false,
			RedirectURIs: []string{"http://127.0.0.1:5000*"},
			WebOrigins:   []string{"http://127.0.0.1:5000/"},
			Attributes: ClientAttributes{
				BackchannnelSessionRequired: "true",
				BackchannnelLogoutURL:       "http://127.0.0.1:5000/callback?logout=backchannel",
			},
		},
		{
			ClientID:     "api_server",
			PublicClient: false,
			RedirectURIs: []string{"http://127.0.0.1:4000*"},
		},
	}
	p.Roles = []string{"supervisor", "administrator", "staff"}
	p.Groups = []string{"A社", "B社", "C社", "D社"}
	return p
}

//
//
//
func main() {

	config := *NewConfig()

	log.Println(config)
	token := getAccessToken(config)
	deleteRealm(config, token)
	_, err := createRealm(config, token)
	if err != nil {
		log.Fatal(err)
	}
	for _, client := range config.Clients {
		client.ProtocolMappers = []ProtocolMapper{*NewProtocolMapper()}
		respBody, err := createClient(config, client, token)
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
	values.Set("username", c.Keycloak.AdminAccount)
	values.Add("password", c.Keycloak.AdminPassword)
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

//
//
//
func createClient(c Config, cl Client, token string) (string, error) {
	return post(c.Keycloak.URL+"/admin/realms/"+c.Realm+"/clients", token, cl)
}

//
//
//
func addRole(c Config, token, role string) (string, error) {
	type RoleInfo struct {
		Name string `json:"name"`
	}
	return post(c.Keycloak.URL+"/admin/realms/"+c.Realm+"/roles", token, RoleInfo{Name: role})
}

//
//
//
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

//
//
//
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
