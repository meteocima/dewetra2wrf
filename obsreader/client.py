'''
Created on May 21, 2020

@author: doy
'''
import requests


class OIDC:

    def __init__(self, **configuration):
        self._token_endpoint = configuration['token_endpoint']
        self._token_introspection_endpoint = \
            configuration['token_introspection_endpoint']
        self._end_session_endpoint = configuration['end_session_endpoint']
        self._userinfo_endpoint = configuration['userinfo_endpoint']

    @classmethod
    def configure(cls, well_known_url):
        response = requests.get(well_known_url)
        if response.status_code != 200:
            raise Exception('Error from well known url')
        configuration = response.json()
        return cls(**configuration)

    def token(self, client_id, username, password):
        params = {
            'grant_type': 'password',
            'client_id': client_id,
            'username': username,
            'password': password,
        }
        response = requests.post(self._token_endpoint, data=params)
        if response.status_code != 200:
            print(response.status_code)
            print(response.content)
            raise Exception('Invalid token request')
        return response.json()

    def userinfo(self, access_token):
        headers = {
            'Authorization': 'Bearer %s' % access_token,
        }
        response = requests.get(self._userinfo_endpoint, headers=headers)
        if response.status_code != 200:
            print(response.content)
            raise Exception('Invalid userinfo request')
        return response.json()

    def introspect(self, client_id, client_secret, token):
        params = {
            'client_id': client_id,
            'client_secret': client_secret,
            'token': token,
        }
        response = requests.post(
            self._token_introspection_endpoint, data=params)
        if response.status_code != 200:
            raise Exception('Invalid introspection request: %s --> %s' %
                            (response.status_code, response.content))
        return response.json()

    def refresh(self, client_id, refresh_token):
        params = {
            'grant_type': 'refresh_token',
            'client_id': client_id,
            'refresh_token': refresh_token,
        }
        response = requests.post(self._token_endpoint, data=params)
        if response.status_code != 200:
            raise Exception('Invalid refresh request')
        return response.json()

    def end_session(self, client_id, refresh_token):
        params = {
            'client_id': client_id,
            'refresh_token': refresh_token,
        }
        response = requests.post(self._end_session_endpoint, data=params)
        if response.status_code != 204:
            raise Exception('Invalid end session request')


if __name__ == '__main__':

    api_url = 'http://172.16.104.16:8020/app'
    auth_wk_url = 'https://testauth.cimafoundation.org/auth/realms/webdrops/.well-known/openid-configuration'
    user = 'andrea.parodi@cimafoundation.org'
    password = '^8J*ITws38Cd4b5Cg*g%iSni!KqMPH'

    dt_from = '202005200000'
    dt_to = '202005210000'

    oidc = OIDC.configure(auth_wk_url)

    auth_r = oidc.token('webdrops', user, password)
    token = auth_r['access_token']

    r = requests.get(f'{api_url}/sensors/list/PLUVIOMETRO/', headers={
        'Authorization': f'Bearer {token}',
        #         'AcrowebRole': 'lexis'
    })

    sensors = r.json()

    print(sensors[0])
    r = requests.get(f'{api_url}/sensors/data/PLUVIOMETRO/{sensors[0]["id"]}/?from={dt_from}&to={dt_to}', headers={
        'Authorization': f'Bearer {token}',
        #         'AcrowebRole': 'lexis'
    })

    data = r.json()

    print(data)
