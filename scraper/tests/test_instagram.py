import pytest
import datetime
from unittest.mock import patch, MagicMock
from scraper.instagram import extract_shortcode, fetch_post_data

def test_extract_shortcode():
    assert extract_shortcode("https://www.instagram.com/p/C_abc123/") == "C_abc123"
    assert extract_shortcode("https://instagram.com/reel/XYZ123?utm_source=ig_web_copy_link") == "XYZ123"
    assert extract_shortcode("https://google.com") is None

@patch("scraper.instagram.instaloader.Post.from_shortcode")
@patch("scraper.instagram.instaloader.Instaloader")
def test_fetch_post_data_success(mock_instaloader, mock_from_shortcode):
    mock_post = MagicMock()
    mock_post.caption = "Delicious recipe caption"
    mock_post.date_utc = datetime.datetime(2023, 1, 1, 12, 0, 0)
    mock_from_shortcode.return_value = mock_post

    caption, date = fetch_post_data("https://www.instagram.com/p/XYZ123/")
    
    assert caption == "Delicious recipe caption"
    assert date == datetime.datetime(2023, 1, 1, 12, 0, 0)
    mock_from_shortcode.assert_called_once()

def test_fetch_post_data_invalid_url():
    with pytest.raises(ValueError, match="Could not extract shortcode"):
        fetch_post_data("https://google.com/")
