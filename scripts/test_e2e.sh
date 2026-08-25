#!/bin/bash
set -e

echo "Starting environment..."
docker-compose up -d

echo "Waiting for services to be ready..."
sleep 10

echo "Sending test scrape request..."
curl -s -X POST http://localhost:8001/scrape \
    -H "Content-Type: application/json" \
    -d '{
        "url": "https://www.instagram.com/p/e2e-test/",
        "raw_text": "Here is an e2e test recipe! \nIngredients: \n- 1 E2E test\n- 2 passes\n\nInstructions:\n1. Run the pipeline.\n2. Observe success."
    }'

echo "Waiting for worker to process..."
sleep 20

echo "Verifying insertion in DB..."
RESULT=$(docker exec socialfoodie-postgres-1 psql -U foodie_user -d socialfoodie -t -c "SELECT COUNT(*) FROM recipes WHERE url = 'https://www.instagram.com/p/e2e-test/';")
RESULT=$(echo $RESULT | xargs) # trim whitespace

if [ "$RESULT" -eq "1" ]; then
    echo "✅ E2E Test Passed: Recipe successfully inserted into DB!"
else
    echo "❌ E2E Test Failed: Recipe not found in DB."
    exit 1
fi

echo "Tearing down environment..."
docker-compose down
