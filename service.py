"""Startup script for this service"""
import asyncio
import logging
import sys

import uvicorn
from pydantic import ValidationError

import exceptions
import tools
from settings import *

if __name__ == '__main__':
    # Read the service settings
    _service_settings = ServiceSettings()
    # Configure the logging module
    logging.basicConfig(
        format='%(levelname)-8s | %(asctime)s | %(name)s | %(message)s',
        level=tools.resolve_log_level(_service_settings.log_level)
    )
    # Log a startup message
    logging.info(f'Starting the {_service_settings.name} service')
    logging.debug('Reading the settings for the AMQP connection')
    try:
        _amqp_settings = AMQPSettings()
    except ValidationError as error:
        logging.error('The settings for the AMQP connection could not be read')
        sys.exit(1)
    logging.debug('Reading the settings for the Service Registry connection')
    try:
        _registry_settings = ServiceRegistrySettings()
    except ValidationError as error:
        logging.error('The settings for the Service Registry could not be read')
        sys.exit(2)
    try:
        _database_settings = DatabaseSettings()
    except ValidationError as error:
        logging.error('The settings for the database connection could not be read')
        sys.exit(3)
    # Get the current event loop
    _loop = asyncio.get_event_loop()
    # Check if the service registry is reachable
    logging.info('Checking the communication to the service registry')
    _registry_available = _loop.run_until_complete(
        tools.is_host_available(
            host=_registry_settings.host,
            port=_registry_settings.port
        )
    )
    if not _registry_available:
        logging.critical(
            'The service registry is not reachable. The service may not be reachable via the '
            'Gateway'
        )
        sys.exit(4)
    else:
        logging.info('SUCCESS: The service registry appears to be running')
    # Check if the message broker is reachable
    logging.info('Checking the communication to the message broker')
    _message_broker_available = _loop.run_until_complete(
        tools.is_host_available(
            host=_amqp_settings.dsn.host,
            port=int(_amqp_settings.dsn.port)
        )
    )
    if not _message_broker_available:
        logging.critical(
            'The message broker is not reachable. Since this is a security issue the service will '
            'not start'
        )
        sys.exit(5)
    else:
        logging.info('SUCCESS: The message registry appears to be running')
    # Check if the database is reachable
    logging.info('Checking the communication to the database')
    _database_available = _loop.run_until_complete(
        tools.is_host_available(
            host=_database_settings.dsn.host,
            port=int(_database_settings.dsn.port)
        )
    )
    if not _database_available:
        logging.critical(
            'The database is not reachable. Since the database stores the necessary geospatial '
            'data served by this service the service cannot start'
        )
        sys.exit(5)
    else:
        logging.info('SUCCESS: The database appears to be running')
    # Validate that the specified layers and resolutions exist and the necessary "geodata" scope
    # exists
    try:
        tools.check_layer_configuration(open('layers.yaml'))
    except exceptions.BadLayerConfigurationError:
        logging.critical('The layer configuration is either malformed or some entries are not in '
                         'the database. Please check your geospatial database and your '
                         'configuration files.')
        sys.exit(6)
    # Start the application
    uvicorn.run(**{
        "app":       "api:geo_data_rest",
        "host":      "0.0.0.0",
        "port":      _service_settings.http_port,
        "log_level": "critical",
        "workers":   1
    })
    
   
