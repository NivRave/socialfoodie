import pika
import os
from scraper.models import ScrapePayload
from scraper.logger import setup_logger

logger = setup_logger(__name__)

RABBITMQ_HOST = os.getenv("RABBITMQ_HOST", "localhost")
RABBITMQ_PORT = int(os.getenv("RABBITMQ_PORT", "5673"))
RABBITMQ_USER = os.getenv("RABBITMQ_USER", "foodie_mq")
RABBITMQ_PASS = os.getenv("RABBITMQ_PASS", "foodie_mq_pass")
QUEUE_NAME = "recipe_scraping_queue"

def get_rabbitmq_connection():
    credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASS)
    parameters = pika.ConnectionParameters(host=RABBITMQ_HOST, port=RABBITMQ_PORT, credentials=credentials)
    return pika.BlockingConnection(parameters)

def publish_scrape_result(payload: ScrapePayload):
    try:
        connection = get_rabbitmq_connection()
        channel = connection.channel()
        
        # Ensure queue exists with DLX configured
        channel.queue_declare(
            queue=QUEUE_NAME, 
            durable=True,
            arguments={
                'x-dead-letter-exchange': 'recipe_dlx'
            }
        )
        
        channel.basic_publish(
            exchange='',
            routing_key=QUEUE_NAME,
            body=payload.model_dump_json(),
            properties=pika.BasicProperties(
                delivery_mode=2,  # make message persistent
            )
        )
        connection.close()
        logger.info(f"Published to {QUEUE_NAME}", extra={"trace_id": payload.trace_id, "queue": QUEUE_NAME})
    except Exception as e:
        logger.error(f"Failed to publish to RabbitMQ: {e}", extra={"trace_id": payload.trace_id}, exc_info=True)
