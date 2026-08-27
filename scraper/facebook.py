import datetime
from typing import Tuple, Optional
import yt_dlp
from scraper.logger import setup_logger

logger = setup_logger(__name__)

def fetch_post_data(url: str) -> Tuple[Optional[str], Optional[datetime.datetime]]:
    try:
        url_str = str(url)
        ydl_opts = {'quiet': True, 'no_warnings': True, 'extract_flat': False}
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url_str, download=False)
            caption = info.get('description')
            timestamp_unix = info.get('timestamp')
            
            date_utc = None
            if timestamp_unix:
                date_utc = datetime.datetime.fromtimestamp(timestamp_unix, tz=datetime.timezone.utc)
                
            return caption, date_utc
    except Exception as e:
        logger.error(f"Failed to fetch facebook post {url}: {e}", extra={"url": url}, exc_info=True)
        raise ValueError(f"Could not extract post data from Facebook URL: {url}") from e
