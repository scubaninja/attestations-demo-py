import requests

__version__ = "0.1.0"


def get_ip_address():
    """
    Fetch the current IP address using httpbin.org.
    
    Returns:
        str: The IP address or an error message
    """
    try:
        response = requests.get('https://httpbin.org/get', timeout=5)
        response.raise_for_status()
        return response.json()['origin']
    except requests.exceptions.RequestException as e:
        return f"Error fetching IP: {e}"


def main():
    """Main entry point for the package."""
    ip_address = get_ip_address()
    print(f"Hello, your IP address is: {ip_address}")
