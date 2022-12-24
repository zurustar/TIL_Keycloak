
#
# docker run -p 8080:8080 -e KEYCLOAK_ADMIN=admin -e KEYCLOAK_ADMIN_PASSWORD=admin quay.io/keycloak/keycloak:20.0.2 start-dev
#

import sys
import requests 
import json
import csv


realm = "corpSite"
base_url = "http://localhost:8080"


def main(filename):
	token = getToken()
	print(token)
	getRealm(token)
	deleteRealm(token, realm)
	getRealm(token)
	createRealm(token, realm)
	getRealm(token)
	for grp in ['grpA','grpB','grpO']:
		createGroup(token, realm, grp)
	getGroups(token, realm)
	with open(filename, "r") as fp:
		f = csv.reader(fp, delimiter=",")
		for row in f:
			createUser(token, realm, row)

def getToken():
	data = {
		'username': 'admin',
		'password': 'admin',
		'grant_type': 'password',
		'client_id': 'admin-cli' 
	}
	resp = requests.post(base_url + '/realms/master/protocol/openid-connect/token', data)
	return json.loads(resp.text)["access_token"]

def deleteRealm(token, realm):
	hdrs = {'Authorization': 'Bearer ' + token}
	resp = requests.delete(base_url + '/admin/realms/' + realm, headers=hdrs)
	if resp.status_code < 200 or resp.status_code >= 300:
		print(resp.text)

def createRealm(token, realm):
	print("createRealm(token, ", realm, ")")
	hdrs = {'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json'}
	data = {'realm': realm, 'enabled': True}
	resp = requests.post(base_url + '/admin/realms', headers=hdrs, data=json.dumps(data))
	if resp.status_code < 200 or resp.status_code >= 300:
		print(resp.text)

def getRealm(token):
	print("getRealm(token)")
	hdrs = {'Authorization': 'Bearer ' + token}
	resp = requests.get(base_url + '/admin/realms', headers=hdrs)
	if resp.status_code < 200 or resp.status_code >= 300:
		print(resp.text)
		return
	data = json.loads(resp.text)
	for realm in data:
		print(realm['realm'])

def createGroup(token, realm, group):
	hdrs = {'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json'}
	data = {'name': group}
	resp = requests.post(base_url + '/admin/realms/' + realm + '/groups', headers=hdrs, data=json.dumps(data))
	if resp.status_code < 200 or resp.status_code >= 300:
		print(resp.text)

def getGroups(token, realm):
	hdrs = {'Authorization': 'Bearer ' + token}
	resp = requests.get(base_url + '/admin/realms/' + realm + '/groups' , headers=hdrs)
	if resp.status_code < 200 or resp.status_code >= 300:
		print(resp.text)
		return
	data = json.loads(resp.text)
	for grp in data:
		print(grp)

def createUser(token, realm, row):
	hdrs = {'Authorization': 'Bearer ' + token}
	grp = row[6]
	mail = row[7]
	print(grp, mail)


if __name__ == '__main__':
	main(sys.argv[1])

