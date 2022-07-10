
import requests
import json
import csv

url = 'http://localhost:8080'
my_realm = 'jikken'
my_roles =  ["jikken_user", "jikken_superuser", "jikken_administrator"]
my_groups =  ["jikken_teamA", "jikken_teamB", "jikken_teamC", "jikken_teamD"]

#   ダミーのユーザとしてこの素晴らしいサイトで生成したデータを試用する。
#     https://testdata.userlocal.jp/
csv_file = './dummy.csv'

#
# アクセストークンを取得
#
def get_access_token():
    print("get_access_token()")
    r = requests.post(
        url+'/realms/master/protocol/openid-connect/token',
        data={'username':'admin', 'password':'admin', 'grant_type': 'password', 'client_id':'admin-cli'})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()
    j = json.loads(r.text)
    return j['access_token']

#
# いま設定されているレルムを取得してみる
#
def get_realms(token):
    print("get_realms("+token+")")
    r = requests.get(url+'/admin/realms',headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()
    return json.loads(r.text)

#
# いまから作ろうとしているレルムがすでにあったら消してみる
#
def del_realms(token, realm):
    print("del_realms("+token+", "+realm+")")
    r = requests.delete(url+'/admin/realms/' + realm, headers={'Authorization': 'Bearer ' +token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()

#
# レルム作成
#
def create_realm(token, realm):
    print("create_realms("+token+", "+realm+")")
    r = requests.post(
        url + '/admin/realms',
        headers={'Authorization': 'Bearer ' + token},
        json={'realm': realm, 'enabled': True})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()

#
# ロール作成
#
def create_role(token, realm, role):
    print("create_role("+token+", "+realm+", "+role+")")
    r = requests.post(
        url+'/admin/realms/' + realm + '/roles',
        headers={'Authorization': 'Bearer ' + token},
        json={'name': role})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()

#
# ロール情報取得
#
def get_roles_info(token, realm):
    print("get_roles_info("+token+", "+realm+")")
    r = requests.get(
        url+'/admin/realms/' + realm + '/roles',
        headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()
    result = {}
    for role in json.loads(r.text):
        result[role['name']]=role
    return result

#
# グループを作る
#
def create_group(token, realm, group):
    print("create_group("+token+", "+realm+", "+group+")")
    r = requests.post(
        url+'/admin/realms/' + realm + '/groups',
        headers={'Authorization': 'Bearer ' + token},
        json={'name': group})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        quit()

#
# ユーザを作成
#   TODO: 同時にロールを追加したい。
#
def create_user(token, realm, user, group):
    print("create_user("+token+", "+realm+", "+user[0]+", "+group+")")
    ary = user[1].split(' ')
    r = requests.post(
        url+'/admin/realms/' + realm + '/users',
        headers={'Authorization': 'Bearer ' + token},
        json={
            'username': user[0],
            'email': user[6],
            'firstName': ary[0],
            'lastName': ary[1],
            'groups': [group],
            'attributes': {"age": user[2], "zipcode": user[9], "address": user[10], "company": user[11]},
            'enabled': True})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# ユーザ情報取得
#
def get_user_info(token, realm, username):
    print("get_user_info("+token+", "+realm+", "+username+")")
    r = requests.get(
        url+'/admin/realms/' + realm + '/users?username='+username ,
        headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    return json.loads(r.text)[0]

#
# ユーザのグループへの追加
#
def set_user_role(token, realm, uid, role_id, role_name):
    print("set_user_role("+token+", "+realm+", "+uid+", "+ role_id+", "+role_name+")")
    r = requests.post(
        url+'/admin/realms/' + realm + '/users/' + uid + "/role-mappings/realm",
        headers={'Authorization': 'Bearer ' + token},
        json=[{'id':role_id,'name': role_name, 'composite': False, 'clientRole': False, 'containerId': ''}])
    if not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()


def main():
    # トークン取得
    token = get_access_token()
    # 現在のレルム一覧を取得
    realms = get_realms(token)
    # もし今から作ろうとしているレルムがすでにあったら削除
    for realm in realms:
        if realm['realm'] == my_realm:
            del_realms(token, my_realm)
    # レルム作成
    create_realm(token, my_realm)
    # レルムロール作成
    for role in my_roles:
        create_role(token, my_realm, role)
    # ロールの情報を取得
    role_info = get_roles_info(token, my_realm)
    # グループ作成
    for g in my_groups:
        create_group(token, my_realm, g)
    # ユーザ情報が書いてあるCSVファイルを一行ずつ読んで追加していく
    text = open(csv_file, "r", encoding="utf-8", errors="", newline="" )
    f = csv.reader(text, delimiter=",", doublequote=True, lineterminator="\r\n", quotechar='"', skipinitialspace=True)
    header = next(f)
    for i, user in enumerate(f):
        # ユーザ作成
        create_user(token, my_realm, user, my_groups[i%len(my_groups)])
        # ユーザ情報の取得
        user = get_user_info(token, my_realm, user[0])
        # このユーザに追加したいロールを決めて、、
        rolename = my_roles[i%len(my_roles)]
        # ユーザにロールをマッピング
        set_user_role(token, my_realm, user['id'], role_info[role]['id'], role)

if __name__ == '__main__':
    main()
