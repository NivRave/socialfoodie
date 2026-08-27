import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock

# Important: Mock the pika connection before importing main
with patch("pika.BlockingConnection"):
    from scraper.main import app

client = TestClient(app)

def test_scrape_endpoint_success():
    payload = {
        "url": "https://www.instagram.com/p/123456789/",
        "raw_text": "Here is a recipe."
    }
    
    with patch("scraper.main.publish_scrape_result") as mock_publish:
        response = client.post("/scrape", json=payload)
        
        assert response.status_code == 202
        data = response.json()
        assert data["message"] == "Scraping task accepted"
        assert "trace_id" in data
        assert data["url"] == payload["url"]
        
        mock_publish.assert_called_once()
        args, kwargs = mock_publish.call_args
        assert str(args[0].source_url) == payload["url"]
        assert args[0].raw_caption == payload["raw_text"]
        assert args[0].platform == "instagram"

def test_scrape_endpoint_facebook():
    payload = {
        "url": "https://www.facebook.com/watch/?v=123",
        "raw_text": "FB recipe."
    }

    with patch("scraper.main.publish_scrape_result") as mock_publish:
        response = client.post("/scrape", json=payload)
        assert response.status_code == 202

        mock_publish.assert_called_once()
        args, kwargs = mock_publish.call_args
        assert args[0].platform == "facebook"

def test_scrape_endpoint_invalid_url():
    payload = {
        "url": "not-a-url"
    }
    
    response = client.post("/scrape", json=payload)
    
    # FastAPI automatically rejects invalid HttpUrl fields
    assert response.status_code == 422
