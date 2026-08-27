import logging
from pythonjsonlogger import jsonlogger

def setup_logger(name: str = "socialfoodie.scraper") -> logging.Logger:
    logger = logging.getLogger(name)
    
    # Avoid adding multiple handlers if setup is called multiple times
    if not logger.handlers:
        handler = logging.StreamHandler()
        formatter = jsonlogger.JsonFormatter(
            '%(asctime)s %(levelname)s %(name)s %(message)s'
        )
        handler.setFormatter(formatter)
        logger.addHandler(handler)
        logger.setLevel(logging.INFO)
        
    return logger
