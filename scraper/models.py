from pydantic import BaseModel, HttpUrl
from typing import Optional
import datetime

class ScrapeRequest(BaseModel):
    url: HttpUrl
    raw_text: Optional[str] = None

class ScrapePayload(BaseModel):
    trace_id: str
    source_url: HttpUrl
    platform: str
    raw_caption: Optional[str] = None
    timestamp: Optional[datetime.datetime] = None
    scraped_at: datetime.datetime
