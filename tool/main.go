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

const baseURL = "http://localhost:8080"

//
//
//
func main() {
	myRealm := "jikken"
	token := getAccessToken()
	log.Println(token)
	delRealm(myRealm, token)
	for _, role := range []string{"jikken_user", "jikken_superuser", "jikken_administrator"} {
		createRole(myRealm, token, role)
	}
	for _, group := range[]string{"jikken_teamA", "jikken_teamB", "jikken_teamC", "jikken_teamD"}{
		createGroup(myRealm, token, group)
	}


}

//
//
//
func getAccessToken() string {
	values := url.Values{}
	values.Set("username", "admin")
	values.Add("password", "admin")
	values.Add("grant_type", "password")
	values.Add("client_id", "admin-cli")
	req, err := http.NewRequest(
		"POST",
		baseURL+"/realms/master/protocol/openid-connect/token",
		strings.NewReader(values.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		panic(fmt.Errorf("%d", resp.StatusCode))
	}
	type Token struct {
		Token string `json:"access_token"`
	}
	var token Token
	err = json.Unmarshal(body, &token)
	if err != nil {
		panic(err)
	}
	return token.Token
}

//
//
//
func delRealm(realm, token string) string {
	return del("/admin/realms/"+realm, token)
}

//
//
//
func createRealm(realm, token string) string {
	type RealmInfo struct {
		Realm string `json:"realm"`
		Enabled bool `json:"enabled"`
	}
	body :=RealmInfo{Realm: realm, Enabled: true}
	return post("/admin/realms/", token, body)
}

//
//
//
func createRole(realm, token, role string) string {
	type RoleInfo struct {
		Name string `json:"name"`
	}
	body :=RoleInfo{Name: role}
	return post("/admin/realms/" + realm + "/roles", token, body)
}

//
//
//
func createGroup(realm, token, group string) string {
	type GroupInfo struct {
		Name string `json:"name"`
	}
	body :=GroupInfo{Name: group}
	return post("/admin/realms/" + realm + "/groups", token, body)
}

//
//
//
func post(path, token string, jsondata any) string {
	buf, err := json.Marshal(jsondata)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(
		"POST", baseURL+path, bytes.NewBuffer(buf))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		panic(fmt.Errorf("%d", resp.StatusCode))
	}
	return string(body)
}


func del(path, token string) string {
	req, err := http.NewRequest(
		"DELETE", baseURL+path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		log.Println(resp.StatusCode)
		log.Println(string(body))
		panic(fmt.Errorf("%d", resp.StatusCode))
	}
	return string(body)
}
