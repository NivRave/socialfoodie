FROM python:3.12-slim

# Create a non-root user and group
RUN groupadd -r appgroup && useradd -r -g appgroup appuser

WORKDIR /app

# Install dependencies (optimizing layer caching)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code into a package directory
COPY . /app/scraper

# Fix ownership of the app directory
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser:appgroup

# Ensure the module can be found
ENV PYTHONPATH=/app

EXPOSE 8001

# Start the uvicorn server
CMD ["uvicorn", "scraper.main:app", "--host", "0.0.0.0", "--port", "8001"]
