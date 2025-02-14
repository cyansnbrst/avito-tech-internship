import requests

REGISTER_URL = "http://localhost:8080/api/auth"

def register_user(username, password):
    payload = {
        "username": username,
        "password": password
    }
    response = requests.post(REGISTER_URL, json=payload)
    if response.status_code == 200:
        return response.json()  
    else:
        return None

users = []
for i in range(1000):
    username = f"user_{i}"
    password = "password"
    token = register_user(username, password)
    if token:
        users.append({"username": username, "token": token["access_token"]})

with open("users.txt", "w") as f:
    for user in users:
        f.write(f"{user['username']} {user['token']}\n")