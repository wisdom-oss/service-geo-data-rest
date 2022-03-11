"""Module for organizing database related functions and operations"""
from sqlalchemy import create_engine
from sqlalchemy.future import Engine

from settings import DatabaseSettings

# Always read the database settings
_settings = DatabaseSettings()

# Create a new database engine from the settings
_engine = create_engine(
    _settings.dsn,
    pool_size=10,
    pool_recycle=90,
    pool_pre_ping=True
)
        

def engine() -> Engine:
    """Access the currently used engine
    
    :return: The currently used engine
    """
    return _engine
