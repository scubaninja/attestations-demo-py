import os
import requests
import feedparser
import tweepy
from datetime import datetime

def get_latest_blogs(rss_url):
    feed = feedparser.parse(rss_url)
    return [
        f"* [{entry.title}]({entry.link})" 
        for entry in feed.entries[:3]
    ]

def get_latest_tweets(api_key):
    client = tweepy.Client(bearer_token=api_key)
    tweets = client.get_users_tweets(id='YOUR_USER_ID', max_results=3)
    return [
        f"* {tweet.text}" 
        for tweet in tweets.data
    ]

def get_github_contributions():
    token = os.getenv('GITHUB_TOKEN')
    headers = {'Authorization': f'token {token}'}
    response = requests.get(
        'https://api.github.com/users/YOUR_USERNAME/events',
        headers=headers
    )
    events = response.json()[:5]
    return [
        f"* [{event['repo']['name']}]({event['repo']['url']})" 
        for event in events
    ]

def update_readme():
    with open('README.md', 'r') as file:
        content = file.read()

    # Update sections
    blogs = '\n'.join(get_latest_blogs(os.getenv('BLOG_RSS_FEED')))
    tweets = '\n'.join(get_latest_tweets(os.getenv('TWITTER_API_KEY')))
    contributions = '\n'.join(get_github_contributions())

    # Replace content between markers
    content = content.replace(
        '<!-- BLOG-POST-LIST:START -->',
        f'<!-- BLOG-POST-LIST:START -->\n{blogs}'
    )
    content = content.replace(
        '<!-- TWITTER:START -->',
        f'<!-- TWITTER:START -->\n{tweets}'
    )
    content = content.replace(
        '<!-- GITHUB-ACTIVITY:START -->',
        f'<!-- GITHUB-ACTIVITY:START -->\n{contributions}'
    )

    with open('README.md', 'w') as file:
        file.write(content)

if __name__ == '__main__':
    update_readme()