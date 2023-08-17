package main

import (
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

const hostport = "127.0.0.1:8080"

type LoginInfo struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not_before_policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

type User struct {
	Name string `json:"name"`
	MailAddress string `json:"mail_address"`
}

func main() {
	log.SetFlags(log.Lshortfile)
	l := getToken()
	realm := "sample-realm"

	// 取得する
	realms, _ := getRealm(l)
	for _, r := range realms {
		if r == realm {
			log.Println("すでにレノムがあったので削除します、レノムの場合は名前を指定する")
			deleteRealm(realm, l)
		}
	}

	// レルムを作る
	createRealm(realm, l)
	log.Println("レノム情報を取得します")
	getRealm(l)

	// グループを作る
	createGroup(realm, "content-editor", l)
	createGroup(realm, "system-administrator", l)
	createGroup(realm, "gomi", l)
	log.Println("グループ情報を取得します")
	groups, _ := getGroup(realm, l)

	// ロールを作る
	createRole(realm, "tech-staff", l)
	createRole(realm, "support-staff", l)
	log.Println("ロール情報を取得します")
	getRole(realm, l)

	// ゴミのグループを消す
	log.Println("グループ情報を削除します、グループの場合はIDを指定する")
	deleteGroup(realm, groups["gomi"], l)
	// 再度取得
	getGroup(realm, l)

	fp, err := os.Open(os.Args[1])
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()

	reader :=csv.NewReader(fp)
	reader.LazyQuotes = true
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		u := User{
			Name: record[0],
			MailAddress: record[6],
		}
		log.Println(u)
		createUser(realm, u)
	}
}

//
// トークンの取得
//
func getToken() LoginInfo {

	form := url.Values{}
	form.Add("client_id", "admin-cli")
	form.Add("username", "admin")
	form.Add("password", "admin")
	form.Add("grant_type", "password")

	cl := &http.Client{}
	u := "http://" + hostport + "/realms/master/protocol/openid-connect/token"
	log.Println(u)
	req, err := http.NewRequest("POST", u, strings.NewReader(form.Encode()))
	if err != nil {
		log.Println(err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := cl.Do(req)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("%d", resp.StatusCode)
		panic(err)
	}
	var l LoginInfo
	err = json.Unmarshal(rbody, &l)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	return l
}

//
// レルムを作る
//
func createRealm(realm string, l LoginInfo) {

	type RealmT struct {
		Realm                       string   `json:"realm"`
		DisplayName                 string   `json:"displayName"`
		Enabled                     bool     `json:"enabled"`
		SSLRequired                 string   `json:"sslRequired"`
		RegistrationAllowed         bool     `json:"registrationAllowed"`
		LoginWithEmailAllowed       bool     `json:"loginWithEmailAllowed"`
		DuplicateEmailsAllowed      bool     `json:"duplicateEmailsAllowed"`
		ResetPasswordAllowed        bool     `json:"resetPasswordAllowed"`
		EditUsernameAllowed         bool     `json:"editUsernameAllowed"`
		BruteForceProtected         bool     `json:"bruteForceProtected"`
		InternationalizationEnabled bool     `json:"internationalizationEnabled"`
		SupportedLocales            []string `json:"supportedLocales"`
		DefaultLocale               string   `json:"defaultLocale"`
	}
	var r RealmT
	r.Realm = realm
	r.DisplayName = realm
	r.Enabled = true
	r.SSLRequired = "external"
	r.RegistrationAllowed = true
	r.LoginWithEmailAllowed = true
	r.DuplicateEmailsAllowed = false
	r.ResetPasswordAllowed = true
	r.EditUsernameAllowed = true
	r.BruteForceProtected = true
	r.InternationalizationEnabled = true
	r.SupportedLocales = []string{"ja", "en"}
	r.DefaultLocale = "ja"

	body, err := postJSON("/admin/realms/", l.AccessToken, r)
	if err != nil {
		panic(err)
	}
	log.Println(body)
}

//
// レノム情報の取得
//
func getRealm(l LoginInfo) ([]string, error) {
	type RealmT struct {
		ID          string `json:"id"`
		Realm       string `json:"realm"`
		DisplayName string `json:"displayName"`
	}
	var realms []RealmT
	body, err := get("/admin/realms", l.AccessToken)
	if err != nil {
		panic(err)
	}
	log.Println(body)
	err = json.Unmarshal([]byte(body), &realms)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Println(realms)
	result := []string{}
	for _, r := range realms {
		result = append(result, r.Realm)
	}
	return result, nil
}

func deleteRealm(realmName string, l LoginInfo) {
	body, err := del("/admin/realms/"+realmName, l.AccessToken)
	if err != nil {
		panic(err)
	}
	log.Println(body)
}

//
// グループを作る
//
func createGroup(realm, grpname string, l LoginInfo) {
	type GroupT struct {
		Name string `json:"name"`
	}
	var g GroupT
	g.Name = grpname
	body, err := postJSON("/admin/realms/"+realm+"/groups", l.AccessToken, g)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Println(body)
}

//
// グループの情報を取得する
//
func getGroup(realm string, l LoginInfo) (map[string]string, error) {
	type GroupT struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Path string `json:"path"`
	}
	var groups []GroupT
	body, err := get("/admin/realms/"+realm+"/groups", l.AccessToken)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(body), &groups)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Println(groups)
	result := map[string]string{}
	for _, g := range groups {
		result[g.Name] = g.ID
	}
	return result, nil
}

func deleteGroup(realmName, groupName string, l LoginInfo) {
	body, err := del("/admin/realms/"+realmName+"/groups/"+groupName, l.AccessToken)
	if err != nil {
		panic(err)
	}
	log.Println(body)
}

//
// ロールを作る
//
func createRole(realm, roleName string, l LoginInfo) {
	type RoleT struct {
		Name string `json:"name"`
	}
	var r RoleT
	r.Name = roleName
	body, err := postJSON("/admin/realms/"+realm+"/roles", l.AccessToken, r)
	if err != nil {
		panic(err)
	}
	log.Println(body)
}

//
// ロールの情報を取得する
//
func getRole(realm string, l LoginInfo) {
	type RoleT struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Composite   bool   `json:"composite"`
		ClientRole  bool   `json:"clientRole"`
		ContainerID string `json:"containerId"`
	}
	var roles []RoleT
	body, err := get("/admin/realms/"+realm+"/roles", l.AccessToken)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(body), &roles)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Println(roles)
}

//
// 汎用関数、、与えられたデータをJSONに変換してPOSTする
//
func postJSON(path, accessToken string, data any) (string, error) {
	cl := &http.Client{}
	b, err := json.Marshal(&data)
	if err != nil {
		log.Println(err)
		return "", err
	}
	u := "http://" + hostport + path
	log.Println("POST " + u)
	req, err := http.NewRequest("POST", u, strings.NewReader(string(b)))
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := cl.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		err = fmt.Errorf("%d", resp.StatusCode)
		return "", err
	}
	return string(rbody), nil
}

//
// 汎用関数、、GETしてくる
//
func get(path, accessToken string) (string, error) {
	cl := &http.Client{}
	u := "http://" + hostport + path
	log.Println("GET " + u)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := cl.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		err = fmt.Errorf("%d", resp.StatusCode)
		return "", err
	}
	return string(rbody), nil
}

//
// 汎用関数、、DELETEしてくる
//
func del(path, accessToken string) (string, error) {
	cl := &http.Client{}
	u := "http://" + hostport + path
	log.Println("DELETE " + u)
	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		log.Println(err)
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := cl.Do(req)
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		err = fmt.Errorf("%d", resp.StatusCode)
		return "", err
	}
	return string(rbody), nil
}

func createUser(u User) () {

}