import requests

proxies = {
  'http': 'http://localhost:7777',
  'https': 'http://localhost:7777',
}

response = requests.get('http://yt-rpc.onrender.com/', proxies=proxies)
print(response.text)