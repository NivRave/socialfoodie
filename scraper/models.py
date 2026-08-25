from pydantic import BaseModel
from typing import Optional
import datetime

class ScrapeRequest(BaseModel):
    url: str

class ScrapePayload(BaseModel):
    trace_id: str
    source_url: str
    raw_caption: Optional[str]
    timestamp: Optional[datetime.datetime]
    scraped_at: datetime.datetime
