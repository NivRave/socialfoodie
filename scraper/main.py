import uuid
import datetime
from fastapi import FastAPI, HTTPException, BackgroundTasks
from scraper.models import ScrapeRequest, ScrapePayload
from scraper.instagram import fetch_post_data
from scraper.publisher import publish_scrape_result
from scraper.logger import setup_logger

logger = setup_logger(__name__)

app = FastAPI(title="SocialFoodie Scraper API")

def scrape_and_publish(trace_id: str, url: str, raw_text: str = None):
    logger.info(f"Starting scrape for {url}", extra={"trace_id": trace_id, "url": url})
    try:
        url_str = str(url)
        platform = "unknown"
        if "instagram.com" in url_str:
            platform = "instagram"
        elif "facebook.com" in url_str or "fb.watch" in url_str:
            platform = "facebook"
            
        if raw_text:
            caption = raw_text
            timestamp = None
        else:
            if platform == "instagram":
                caption, timestamp = fetch_post_data(url) # This is instagram
            elif platform == "facebook":
                from scraper.facebook import fetch_post_data as fetch_fb_data
                caption, timestamp = fetch_fb_data(url)
            else:
                raise ValueError(f"Unsupported platform for URL: {url}")

        payload = ScrapePayload(
            trace_id=trace_id,
            source_url=str(url),
            platform=platform,
            raw_caption=caption,
            timestamp=timestamp,
            scraped_at=datetime.datetime.now(datetime.timezone.utc)
        )
        publish_scrape_result(payload)
    except Exception as e:
        logger.error(f"Scraping task failed: {e}", extra={"trace_id": trace_id}, exc_info=True)

@app.post("/scrape", status_code=202)
async def scrape_instagram_post(request: ScrapeRequest, background_tasks: BackgroundTasks):
    trace_id = str(uuid.uuid4())
    
    # Schedule the actual scraping and publishing in the background
    background_tasks.add_task(scrape_and_publish, trace_id, request.url, request.raw_text)
    
    return {
        "message": "Scraping task accepted",
        "trace_id": trace_id,
        "url": request.url
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("scraper.main:app", host="0.0.0.0", port=8001, reload=True)
