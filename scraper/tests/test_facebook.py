import pytest
import datetime
from unittest.mock import patch, MagicMock
from scraper.facebook import fetch_post_data

def test_fetch_post_data_success():
    mock_info = {
        "description": "This is a delicious recipe!",
        "timestamp": 1704110400 # 2024-01-01 12:00:00 UTC
    }
    
    mock_ydl = MagicMock()
    mock_ydl.extract_info.return_value = mock_info
    
    with patch("yt_dlp.YoutubeDL") as mock_ydl_class:
        mock_ydl_class.return_value.__enter__.return_value = mock_ydl
        caption, timestamp = fetch_post_data("https://www.facebook.com/watch/?v=12345")
        
        assert caption == "This is a delicious recipe!"
        assert timestamp == datetime.datetime(2024, 1, 1, 12, 0, tzinfo=datetime.timezone.utc)

def test_fetch_post_data_exception():
    mock_ydl = MagicMock()
    mock_ydl.extract_info.side_effect = Exception("Network error")

    with patch("yt_dlp.YoutubeDL") as mock_ydl_class:
        mock_ydl_class.return_value.__enter__.return_value = mock_ydl
        with pytest.raises(ValueError, match="Could not extract post data from Facebook URL"):
            fetch_post_data("https://www.facebook.com/watch/?v=error")
