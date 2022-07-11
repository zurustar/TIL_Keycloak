
import requests
import json
import csv
import time

url = 'http://localhost:8080'
my_realm = 'jikken'
my_roles =  ["jikken_user", "jikken_superuser", "jikken_administrator"]
my_groups =  ["jikken_teamA", "jikken_teamB", "jikken_teamC", "jikken_teamD"]

my_client = "demo_reverse_proxy"

#   ダミーのユーザとしてこの素晴らしいサイトで生成したデータを試用する。
#     https://testdata.userlocal.jp/
csv_file = './dummy.csv'

#
# アクセストークンを取得
#
def get_access_token():
    r = requests.post(
        url+'/realms/master/protocol/openid-connect/token',
        data={'username':'admin', 'password':'admin', 'grant_type': 'password', 'client_id':'admin-cli'})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    j = json.loads(r.text)
    return j['access_token']

#
# いま設定されているレルムを取得してみる
#
def get_realms(token):
    r = requests.get(url+'/admin/realms',headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    return json.loads(r.text)

#
# いまから作ろうとしているレルムがすでにあったら消してみる
#
def del_realms(token, realm):
    r = requests.delete(url+'/admin/realms/' + realm, headers={'Authorization': 'Bearer ' +token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# レルム作成
#
def create_realm(token, realm):
    r = requests.post(
        url + '/admin/realms',
        headers={'Authorization': 'Bearer ' + token},
        json={'realm': realm, 'enabled': True})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# ロール作成
#
def create_role(token, realm, role):
    # とりあえず作る
    r = requests.post(
        url+'/admin/realms/' + realm + '/roles',
        headers={'Authorization': 'Bearer ' + token},
        json={'name': role})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

def get_roles(token, realm):
    # 作った情報を取得
    r = requests.get(
        url+'/admin/realms/' + realm + '/roles',
        headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    return json.loads(r.text)

#
# グループを作る
#
def create_group(token, realm, group):
    # 作る
    r = requests.post(
        url+'/admin/realms/' + realm + '/groups',
        headers={'Authorization': 'Bearer ' + token},
        json={'name': group})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

def get_groups(token, realm):
    # 作ったグループの情報を取得する
    r = requests.get(
        url+'/admin/realms/' + realm + '/groups',
        headers={'Authorization': 'Bearer ' + token})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    return json.loads(r.text)

#
# ユーザを作成
#
def create_user(token, realm, user, group):
    ary = user[1].split(' ')
    r = requests.post(
        url+'/admin/realms/' + realm + '/users',
        headers={'Authorization': 'Bearer ' + token},
        json={
            'username': user[0],
            'email': user[6],
            'firstName': ary[0],
            'lastName': ary[1],
            'groups': [group], # グループに登録できる、いずれ別APIに分離する予定
            'attributes': {
                "age": user[2],
                "zipcode": user[9],
                "address": user[10],
                "company": user[11]}, # 任意の情報を保存できる
            'enabled': True})
    if  not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# ユーザ情報取得
#
def get_user(token, realm, username):
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
    r = requests.post(
        url+'/admin/realms/' + realm + '/users/' + uid + "/role-mappings/realm",
        headers={'Authorization': 'Bearer ' + token},
        json=[{'id':role_id,'name': role_name, 'composite': False, 'clientRole': False, 'containerId': ''}])
    if not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# クライアントの作成
#
def create_client(token, realm, client):
    r = requests.post(
        url+'/admin/realms/' + realm + '/clients',
        headers={'Authorization': 'Bearer ' + token},
        json={'clientId': client})
    if not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()

#
# クライアント情報の取得
#
def get_client(token, realm, client):
    r = requests.get(
        url+'/admin/realms/' + my_realm + '/clients?clientId=' + client,
        headers={'Authorization': 'Bearer ' + token})
    if not (200 <= r.status_code and r.status_code < 300):
        print(r.status_code)
        print(r.text)
        quit()
    return json.loads(r.text)




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
    role2id = {}
    for role in my_roles:
        r = create_role(token, my_realm, role)
    roles = get_roles(token, my_realm)
    for role in roles:
        role2id[role['name']] = role['id']
    print(role2id)

    # グループ作成
    for g in my_groups:
        create_group(token, my_realm, g)
    print(get_groups(token, my_realm))

    # ユーザ情報が書いてあるCSVファイルを一行ずつ読んで追加していく
    text = open(csv_file, "r", encoding="utf-8", errors="", newline="" )
    f = csv.reader(text, delimiter=",", doublequote=True, lineterminator="\r\n", quotechar='"', skipinitialspace=True)
    header = next(f)
    for i, user in enumerate(f):
        print(user)
        # ユーザ作成
        create_user(token, my_realm, user, my_groups[i%len(my_groups)])
        # ユーザ情報の取得
        user = get_user(token, my_realm, user[0])
        # このユーザに追加したいロールを決めて、、
        rolename = my_roles[i%len(my_roles)]
        # ユーザにロールをマッピング
        set_user_role(token, my_realm, user['id'], role2id[rolename], rolename)

    # レルムにクライアントを登録
    create_client(token, my_realm, my_client)
    # クライアントの情報を取得　※特定のクライアントの情報を取得することはできないか？
    client = get_client(token, my_realm, my_client)
    print(client)



if __name__ == '__main__':
    start = time.time()
    main()
    print(time.time() - start)