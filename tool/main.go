package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const baseURL = "http://localhost:8080"

//
//
//
func main() {
	log.SetFlags(log.Lshortfile)

	myRealm := "jikken"
	token := getAccessToken()
	log.Println(token)
	for _, realm := range getRealms(token) {
		log.Println("check realm", realm)
		if realm == myRealm {
			log.Println("delete realm", realm)
			delRealm(myRealm, token)
		}
	}
	log.Println("create realm", myRealm)
	createRealm(myRealm, token)
	for _, role := range []string{"jikken_user", "jikken_superuser", "jikken_administrator"} {
		log.Println("create role", role)
		createRole(myRealm, token, role)
	}
	roles := getRoles(myRealm, token)
	for _, group := range []string{"jikken_teamA", "jikken_teamB", "jikken_teamC", "jikken_teamD"} {
		log.Println("create group", group)
		createGroup(myRealm, token, group)
	}
	groups := getGroups(myRealm, token)

	fp, err := os.Open("./dummy.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	r := csv.NewReader(fp)
	i := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if i != 0 {
			_, err := createUser(myRealm, token, record)
			if err != nil {
				user := getUser(myRealm, token, record[0])

				// 以下のグループの追加とロールの追加は現状うまくいっていない
				err = setUserGroup(myRealm, token, user, groups[i%len(groups)])
				if err != nil {
					err = setUserRole(myRealm, token, user, roles[i%len(roles)])
				}
			}
		}
		i += 1
	}
	clientId := "demo_reverse_proxy"
	createClient(myRealm, token, clientId)
	clinfo := getClient(myRealm, token, clientId)
	log.Println(string(clinfo))
}

// **************************************************************************
//
// トークンの取得
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

// **************************************************************************
//
//　レルムの操作
//
type RealmInfo struct {
	Realm   string `json:"realm"`
	Enabled bool   `json:"enabled"`
}

// レルム情報取得
func getRealms(token string) []string {
	b := get("/admin/realms", token)
	var realms []RealmInfo
	err := json.Unmarshal(b, &realms)
	if err != nil {
		log.Fatal(err)
	}
	result := []string{}
	for _, r := range realms {
		result = append(result, r.Realm)
	}
	return result
}

// レルム削除
func delRealm(realm, token string) string {
	return del("/admin/realms/"+realm, token)
}

// レルム作成
func createRealm(realm, token string) (string, error) {
	body := RealmInfo{Realm: realm, Enabled: true}
	return post("/admin/realms/", token, body)
}

// **************************************************************************
//
// ロールの操作
//
type RoleInfo struct {
	Name string `json:"name"`
}
type RoleInfoDetail struct {
	RoleInfo
	ID string `json:"id"`
}

// ロール情報取得
func getRoles(realm, token string) []RoleInfoDetail {
	b := get("/admin/realms/"+realm+"/roles", token)
	var roles []RoleInfoDetail
	err := json.Unmarshal(b, &roles)
	if err != nil {
		log.Fatal(err)
	}
	return roles
}

// ロール作成
func createRole(realm, token, role string) (string, error) {
	body := RoleInfo{Name: role}
	return post("/admin/realms/"+realm+"/roles", token, body)
}

// **************************************************************************
//
// グループの操作
//
type GroupInfo struct {
	Name string `json:"name"`
}

type GroupInfoDetail struct {
	GroupInfo
	ID string `json:"id"`
}

// グループ情報取得
func getGroups(realm, token string) []GroupInfoDetail {
	b := get("/admin/realms/"+realm+"/groups", token)
	var groups []GroupInfoDetail
	err := json.Unmarshal(b, &groups)
	if err != nil {
		log.Fatal(err)
	}
	return groups
}

// グループ作成
func createGroup(realm, token, group string) (string, error) {
	body := GroupInfo{Name: group}
	return post("/admin/realms/"+realm+"/groups", token, body)
}

// **************************************************************************
//
// ユーザ情報の操作
//

type UserAttributes struct {
	Age     []string `json:"age"`
	ZipCode []string `json:"zipcode"`
	Address []string `json:"address"`
	Company []string `json:"company"`
}

type UserInfo struct {
	Username   string         `json:"username"`
	Email      string         `json:"email"`
	FirstName  string         `json:"firstName"`
	LastName   string         `json:"lastName"`
	Groups     []string       `json:"groups"`
	Attributes UserAttributes `json:"attributes"`
	Enabled    bool           `json:"enabled"`
}

type UserInfoDetail struct {
	UserInfo
	ID string `json:"id"`
}

// ユーザ登録
func createUser(realm, token string, data []string) (string, error) {
	attr := UserAttributes{
		Age:     []string{data[2]},
		ZipCode: []string{data[9]},
		Address: []string{data[10]},
		Company: []string{data[11]}}
	ary := strings.Split(data[0], " ")
	body := UserInfo{
		Username:   data[0],
		Email:      data[6],
		FirstName:  ary[0],
		LastName:   ary[1],
		Groups:     []string{},
		Attributes: attr,
		Enabled:    true}
	return post("/admin/realms/"+realm+"/users", token, body)
}

// ユーザ情報取得
func getUser(realm, token, username string) UserInfoDetail {
	username = url.QueryEscape(username)
	b := get("/admin/realms/"+realm+"/users?username="+username, token)
	userInfoDetail := []UserInfoDetail{}
	err := json.Unmarshal(b, &userInfoDetail)
	if err != nil {
		log.Println(string(b))
		log.Fatal(err)
	}
	return userInfoDetail[0]
}

// 特定のユーザをグループに登録する
func setUserGroup(realm, token string, user UserInfoDetail, group GroupInfoDetail) error {
	user.Groups = []string{group.Name}
	log.Println(user)
	_, err := put("/admin/realms/"+realm+"/users/"+user.ID, token, user)
	return err
}

// 特定のユーザのロールを設定する
func setUserRole(realm, token string, user UserInfoDetail, role RoleInfoDetail) error {
	type UserRole struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Composite   bool   `json:"composite"`
		ClientRole  bool   `json:"clientRole"`
		ContainerID string `json:"containerId"`
	}
	body := []UserRole{{
		ID:          role.ID,
		Name:        role.Name,
		Composite:   false,
		ClientRole:  false,
		ContainerID: "",
	}}
	_, err := post("/admin/realms/"+realm+"/users/"+user.ID+"/role-mappings/realm", token, body)
	return err
}

// **************************************************************************
//
// クライアントの操作
//

// クライアントの登録
func createClient(realm, token, client string) {
	u := "/admin/realms/" + realm + "/clients"
	type ClientInfo struct {
		ClientID string `json:"clientId"`
	}
	body := ClientInfo{ClientID: client}
	post(u, token, body)
}

// クライアント情報の習得
func getClient(realm, token, client string) []byte{
	u := "/admin/realms/" + realm + "/clients?clientId=" + client
	return get(u, token)
}

// **************************************************************************
//
// 参照
//
func get(path, token string) []byte {
	u := baseURL + path
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
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
	return body
}

//
// 登録
//
func post(path, token string, jsondata any) (string, error) {
	buf, err := json.Marshal(jsondata)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(
		"POST", baseURL+path, bytes.NewBuffer(buf))
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
		log.Println(baseURL + path)
		log.Println(string(buf))
		log.Println(string(body))
		return string(body), fmt.Errorf("%d", resp.StatusCode)
	}
	return string(body), nil
}

//
// 更新
//
func put(path, token string, jsondata any) (string, error) {
	buf, err := json.Marshal(jsondata)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req, err := http.NewRequest(
		"PUT", baseURL+path, bytes.NewBuffer(buf))
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

//
// 削除
//
func del(path, token string) string {
	req, err := http.NewRequest(
		"DELETE", baseURL+path, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
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
	return string(body)
}
