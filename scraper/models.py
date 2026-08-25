from pydantic import BaseModel, HttpUrl
from typing import Optional
import datetime

class ScrapeRequest(BaseModel):
    url: HttpUrl
    raw_text: Optional[str] = None

class ScrapePayload(BaseModel):
    trace_id: str
    source_url: str
    raw_caption: Optional[str]
    timestamp: Optional[datetime.datetime]
    scraped_at: datetime.datetime
