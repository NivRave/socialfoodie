import instaloader
import re
import datetime
from typing import Tuple, Optional
from scraper.logger import setup_logger

logger = setup_logger(__name__)

def extract_shortcode(url: str) -> Optional[str]:
    # Match shortcode in https://www.instagram.com/p/XYZ123/ or https://www.instagram.com/reel/XYZ123/
    match = re.search(r'instagram\.com/(?:p|reel)/([^/?#&]+)', url)
    if match:
        return match.group(1)
    return None

def fetch_post_data(url: str) -> Tuple[Optional[str], Optional[datetime.datetime]]:
    shortcode = extract_shortcode(url)
    if not shortcode:
        raise ValueError(f"Could not extract shortcode from URL: {url}")

    L = instaloader.Instaloader()
    # Attempting anonymous scrape
    try:
        post = instaloader.Post.from_shortcode(L.context, shortcode)
        return post.caption, post.date_utc
    except Exception as e:
        logger.error(f"Failed to fetch instagram post {url}: {e}", extra={"url": url}, exc_info=True)
        raise e
