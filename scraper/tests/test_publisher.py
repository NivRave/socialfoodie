import json
import pytest
from unittest.mock import MagicMock, patch
from scraper.publisher import publish_scrape_result
from scraper.models import ScrapePayload
import datetime

def test_publish_scrape_result():
    payload = ScrapePayload(
        trace_id="test-123",
        source_url="https://www.instagram.com/p/test/",
        raw_caption="Recipe text",
        timestamp=None,
        scraped_at=datetime.datetime.now(datetime.timezone.utc)
    )

    mock_connection = MagicMock()
    mock_channel = MagicMock()
    mock_connection.channel.return_value = mock_channel

    with patch("scraper.publisher.get_rabbitmq_connection", return_value=mock_connection):
        publish_scrape_result(payload)

    mock_channel.queue_declare.assert_called_with(queue='recipe_scraping_queue', durable=True)
    
    mock_channel.basic_publish.assert_called_once()
    args, kwargs = mock_channel.basic_publish.call_args
    assert kwargs["exchange"] == ""
    assert kwargs["routing_key"] == "recipe_scraping_queue"
    
    published_body = json.loads(kwargs["body"])
    assert published_body["trace_id"] == "test-123"
    assert published_body["source_url"] == "https://www.instagram.com/p/test/"
    assert published_body["raw_caption"] == "Recipe text"
