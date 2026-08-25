import uuid
import datetime
from fastapi import FastAPI, HTTPException, BackgroundTasks
from scraper.models import ScrapeRequest, ScrapePayload
from scraper.instagram import fetch_post_data
from scraper.publisher import publish_scrape_result

app = FastAPI(title="SocialFoodie Scraper API")

def scrape_and_publish(trace_id: str, url: str, raw_text: str = None):
    print(f"[{trace_id}] Starting scrape for {url}")
    try:
        if raw_text:
            caption = raw_text
            timestamp = None
        else:
            caption, timestamp = fetch_post_data(url)
        payload = ScrapePayload(
            trace_id=trace_id,
            source_url=str(url),
            raw_caption=caption,
            timestamp=timestamp,
            scraped_at=datetime.datetime.now(datetime.timezone.utc)
        )
        publish_scrape_result(payload)
    except Exception as e:
        print(f"[{trace_id}] Scraping task failed: {e}")

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
