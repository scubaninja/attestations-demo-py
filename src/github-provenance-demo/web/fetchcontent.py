import os
import logging
import requests
import feedparser
import tweepy
from datetime import datetime
from typing import List, Optional

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Constants
REQUEST_TIMEOUT = 10  # seconds
MAX_RETRIES = 3


def get_latest_blogs(rss_url: str, max_items: int = 3) -> List[str]:
    """
    Fetch latest blog posts from RSS feed.
    
    Args:
        rss_url: URL of the RSS feed
        max_items: Maximum number of items to return
        
    Returns:
        List of formatted blog post links
    """
    if not rss_url:
        logger.warning("RSS URL not provided")
        return []
    
    try:
        feed = feedparser.parse(rss_url)
        
        if feed.bozo:
            logger.error(f"Invalid RSS feed: {feed.bozo_exception}")
            return []
        
        return [
            f"* [{entry.title}]({entry.link})" 
            for entry in feed.entries[:max_items]
            if hasattr(entry, 'title') and hasattr(entry, 'link')
        ]
    except Exception as e:
        logger.error(f"Error fetching blogs: {e}")
        return []


def get_latest_tweets(api_key: str, user_id: str, max_results: int = 3) -> List[str]:
    """
    Fetch latest tweets from Twitter API.
    
    Args:
        api_key: Twitter API bearer token
        user_id: Twitter user ID
        max_results: Maximum number of tweets to return
        
    Returns:
        List of formatted tweets
    """
    if not api_key or not user_id:
        logger.warning("Twitter API key or user ID not provided")
        return []
    
    try:
        client = tweepy.Client(bearer_token=api_key)
        response = client.get_users_tweets(id=user_id, max_results=max_results)
        
        if not response or not response.data:
            logger.warning("No tweets found")
            return []
        
        return [f"* {tweet.text}" for tweet in response.data]
    except tweepy.TweepyException as e:
        logger.error(f"Twitter API error: {e}")
        return []
    except Exception as e:
        logger.error(f"Error fetching tweets: {e}")
        return []


def get_github_contributions(username: str, max_events: int = 5) -> List[str]:
    """
    Fetch recent GitHub contributions.
    
    Args:
        username: GitHub username
        max_events: Maximum number of events to return
        
    Returns:
        List of formatted GitHub events
    """
    token = os.getenv('GITHUB_TOKEN')
    if not username:
        logger.warning("GitHub username not provided")
        return []
    
    headers = {}
    if token:
        headers['Authorization'] = f'token {token}'
    
    try:
        response = requests.get(
            f'https://api.github.com/users/{username}/events',
            headers=headers,
            timeout=REQUEST_TIMEOUT
        )
        response.raise_for_status()
        
        events = response.json()[:max_events]
        return [
            f"* [{event['repo']['name']}]({event['repo']['url']})" 
            for event in events
            if 'repo' in event and 'name' in event['repo'] and 'url' in event['repo']
        ]
    except requests.exceptions.Timeout:
        logger.error("GitHub API request timed out")
        return []
    except requests.exceptions.RequestException as e:
        logger.error(f"GitHub API error: {e}")
        return []
    except (KeyError, ValueError) as e:
        logger.error(f"Error parsing GitHub response: {e}")
        return []


def update_readme(
    readme_path: str = 'README.md',
    blog_rss_feed: Optional[str] = None,
    twitter_api_key: Optional[str] = None,
    twitter_user_id: Optional[str] = None,
    github_username: Optional[str] = None
) -> bool:
    """
    Update README.md with latest content from various sources.
    
    Args:
        readme_path: Path to README.md file
        blog_rss_feed: RSS feed URL for blogs
        twitter_api_key: Twitter API key
        twitter_user_id: Twitter user ID
        github_username: GitHub username
        
    Returns:
        True if update was successful, False otherwise
    """
    try:
        with open(readme_path, 'r', encoding='utf-8') as file:
            content = file.read()
    except FileNotFoundError:
        logger.error(f"README file not found at {readme_path}")
        return False
    except IOError as e:
        logger.error(f"Error reading README: {e}")
        return False

    # Fetch content with environment variable fallbacks
    blog_rss_feed = blog_rss_feed or os.getenv('BLOG_RSS_FEED')
    twitter_api_key = twitter_api_key or os.getenv('TWITTER_API_KEY')
    twitter_user_id = twitter_user_id or os.getenv('TWITTER_USER_ID')
    github_username = github_username or os.getenv('GITHUB_USERNAME')

    blogs = '\n'.join(get_latest_blogs(blog_rss_feed))
    tweets = '\n'.join(get_latest_tweets(twitter_api_key, twitter_user_id))
    contributions = '\n'.join(get_github_contributions(github_username))

    # Update sections only if markers exist
    sections = {
        '<!-- BLOG-POST-LIST:START -->': blogs,
        '<!-- TWITTER:START -->': tweets,
        '<!-- GITHUB-ACTIVITY:START -->': contributions
    }
    
    for marker, new_content in sections.items():
        if marker in content:
            content = content.replace(marker, f'{marker}\n{new_content}')
        else:
            logger.warning(f"Marker {marker} not found in README")

    try:
        with open(readme_path, 'w', encoding='utf-8') as file:
            file.write(content)
        logger.info("README updated successfully")
        return True
    except IOError as e:
        logger.error(f"Error writing README: {e}")
        return False


if __name__ == '__main__':
    success = update_readme()
    exit(0 if success else 1)