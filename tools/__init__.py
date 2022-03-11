"""A collection of tools used multiple times throughout this service"""
from __future__ import annotations

import asyncio
import logging
import sys
import time
from typing import IO

import sqlalchemy
import ujson
import yaml
from pydantic import ValidationError
from sqlalchemy.engine import Inspector

import database
from exceptions import BadLayerConfigurationError
from models.geo import LayerConfiguration, LayerResolution


def resolve_log_level(level: str) -> int:
    """Resolve the logging level from a string
    
    This method will try to get the actual logging level from the logging package
    
    If no valid logging level is supplied this method will return the info level
    
    :param level: The name of the level which should be resolved
    :return: The logging level which may be used in the configuration of loggers
    """
    return getattr(logging, level.upper(), logging.INFO)


async def is_host_available(
        host: str,
        port: int,
        timeout: int = 10
) -> bool:
    """Check if the specified host is reachable on the specified port

    :param host: The hostname or ip address which shall be checked
    :param port: The port which shall be checked
    :param timeout: Max. duration of the check
    :return: A boolean indicating the status
    """
    _end_time = time.time() + timeout
    while time.time() < _end_time:
        try:
            # Try to open a connection to the specified host and port and wait a maximum time of
            # five seconds
            _s_reader, _s_writer = await asyncio.wait_for(asyncio.open_connection(host, port),
                                                          timeout=5)
            # Close the stream writer again
            _s_writer.close()
            # Wait until the writer is closed
            await _s_writer.wait_closed()
            return True
        except:
            # Since the connection could not be opened wait 5 seconds before trying again
            await asyncio.sleep(5)
    return False


def check_layer_configuration(config_file: bytes | IO[bytes] | str | IO[str]):
    """Check if the configuration of the layer is correct and all layers and resolutions are
    present in the database
    
    :param config_file: The layer configuration
    """
    # Try to read the layer configuration file
    _config = yaml.safe_load(config_file)
    # Try to parse the configuration file into the python objects for those
    _layers = {}
    for layer in _config:
        try:
            _layers.update({layer['name']: LayerConfiguration.parse_obj(layer)})
        except ValidationError:
            logging.error('Unable to read the configuration for the layer: %s', layer['name'])
    logging.info('The layer configuration was successfully read')
    # Now check if the database contains the specified schemas
    db_inspector: Inspector = sqlalchemy.inspect(database.engine())
    for layer_name, layer_config in _layers.items():
        logging.info('Checking the configuration of the layer: %s', layer_name)
        logging.info('Checking if the specified schema exists in the database: %s\\%s',
                     layer_name, layer_config.database_schema)
        if layer_config.database_schema not in db_inspector.get_schema_names():
            logging.error(
                '\u274C The specified schema was not found in the database: %s\\%s',
                layer_name, layer_config.database_schema
            )
            raise BadLayerConfigurationError()
        logging.info('\u2705 The specified schema was found in the database: %s\\%s',
                     layer_name, layer_config.database_schema)
        logging.info('Checking if the specified resolutions exist in the database: %s\\%s',
                     layer_name, layer_config.database_schema)
        for resolution in layer_config.resolutions:
            logging.info('Checking if the resolution "%s" is present in the database: %s\\%s\\%s',
                         resolution.name, layer_name, layer_config.database_schema,
                         resolution.table_name)
            if not db_inspector.has_table(resolution.table_name, layer_config.database_schema):
                logging.error('\u274C The resolution "%s" is not in the database: %s\\%s\\%s',
                              resolution.name, layer_name, layer_config.database_schema,
                              resolution.table_name)
                raise BadLayerConfigurationError()
            logging.info('\u2705 The resolution "%s" is in the database: %s\\%s\\%s',
                         resolution.name, layer_name, layer_config.database_schema,
                         resolution.table_name)
        logging.info('\u2705 All specified resolutions are present in the database: %s\\%s',
                     layer_name, layer_config.database_schema)
            
