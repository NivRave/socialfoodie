import uuid
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

app = FastAPI(title="SocialFoodie Scraper API")

class ScrapeRequest(BaseModel):
    url: str

@app.post("/scrape", status_code=202)
async def scrape_instagram_post(request: ScrapeRequest):
    trace_id = str(uuid.uuid4())
    
    # TODO: Implement scraping using instaloader or other means
    # TODO: Push result to RabbitMQ
    
    return {
        "message": "Scraping task accepted",
        "trace_id": trace_id,
        "url": request.url
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
