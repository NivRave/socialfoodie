FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code into a package directory
COPY . /app/scraper

# Ensure the module can be found
ENV PYTHONPATH=/app

EXPOSE 8001

# Start the uvicorn server
CMD ["uvicorn", "scraper.main:app", "--host", "0.0.0.0", "--port", "8001"]
