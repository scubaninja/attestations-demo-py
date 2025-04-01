import requests

def main():
    print("Hello, your IP address is: " + requests.get('https://httpbin.org/get').json()['origin'])

__version__ = "0.1.0"
