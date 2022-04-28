"""FastAPI Implementation of the REST-Service for receiving GeoJSONs"""
import logging
import uuid
from typing import Optional, Union

import ujson
import yaml
from amqp_rpc_client import Client as RPCClient
from fastapi import FastAPI, Request, Path, HTTPException
from fastapi.params import Query
from py_eureka_client.eureka_client import EurekaClient
from sqlalchemy import text
from starlette.middleware.gzip import GZipMiddleware
from starlette.responses import JSONResponse

import database
from api import security
from models.amqp import TokenIntrospectionRequest
from models.geo import LayerConfiguration
from settings import ServiceSettings, ServiceRegistrySettings, AMQPSettings

geo_data_rest = FastAPI()
geo_data_rest.add_middleware(GZipMiddleware, minimum_size=50)

_logger = logging.getLogger('REST-API')
_map_layers: Optional[dict[str, LayerConfiguration]] = {}
_amqp_client: Optional[RPCClient] = None
_amqp_exchange = AMQPSettings().auth_exchange
_service_registry_client: Optional[EurekaClient] = None


@geo_data_rest.on_event('startup')
async def service_startup():
    """Handle the service startup"""
    # Enable the global setup of the amqp client
    global _map_layers, _service_registry_client, _amqp_client
    # Read the necessary configurations
    _service_settings = ServiceSettings()
    _registry_settings = ServiceRegistrySettings()
    # Register the worker at the service registry
    _service_registry_client = EurekaClient(
        app_name=_service_settings.name,
        eureka_server='http://{}:{}'.format(_registry_settings.host, _registry_settings.port),
        instance_port=_service_settings.http_port,
        should_discover=False,
        should_register=True,
        renewal_interval_in_secs=1,
        duration_in_secs=5
    )
    _service_registry_client.start()
    _service_registry_client.status_update('STARTING')
    # Try to read the layers.yaml
    raw_layer_config = yaml.safe_load(open('layers.yaml'))
    # Create the configurations of the layers
    for layer in raw_layer_config:
        _map_layers.update({layer['name']: LayerConfiguration.parse_obj(layer)})
    # == AMQP Client Setup ==
    # Read the AMQP settings
    _amqp_settings = AMQPSettings()
    # Create the new client
    _amqp_client = RPCClient(_amqp_settings.dsn)
    # == AMQP Client Setup done
    _service_registry_client.status_update('UP')
    _logger.info('API is now ready to accept requests')


@geo_data_rest.on_event('shutdown')
async def handle_shutdown():
    """Handle the service shutdown"""
    global _service_registry_client
    # Deregister the client
    _service_registry_client.stop()


@geo_data_rest.middleware('http')
async def check_user_scope(request: Request, call_next):
    """This middleware will validate the authorization token present in the incoming request for
    the scope that is assigned to it. This validation will be done via AMQP
    
    :param request: The incoming request
    :param call_next: The next thing that should happen
    :return: The response
    """
    # Access the request headers
    headers = request.headers
    # Check for a present request id and use it for logging purposes
    _req_id = headers.get('X-Request-ID', uuid.uuid4().hex)
    _logger.debug('%s:%s - %s - Received new request for geospatial data',
                  request.client.host, request.client.port, _req_id)
    _logger.debug('%s:%s - %s - Checking the request for a valid Bearer Token',
                  request.client.host, request.client.port, _req_id)
    # Check if the headers contain the 'Authorization' header
    _authorization_header: Optional[str] = request.headers.get('Authorization', None)
    if _authorization_header is None:
        _logger.warning('%s:%s - %s - The request did not contain a "Authorization" header. ['
                        'REJECTED REQUEST]',
                        request.client.host, request.client.port, _req_id)
        return JSONResponse(
            status_code=400,
            content={
                "error": "missing_authorization_header"
            }
        )
    _logger.debug('%s:%s - %s - Found the "Authorization" header in the request',
                  request.client.host, request.client.port, _req_id)
    # Check if the header value contains the value "Bearer"
    if not ("Bearer" or "bearer") in _authorization_header:
        _logger.warning('%s:%s - %s - The request did not contain a supported authorization '
                        'method [REJECTED REQUEST]',
                        request.client.host, request.client.port, _req_id)
        return JSONResponse(
            status_code=400,
            content={
                "error": "unsupported_authorization_method"
            }
        )
    # Remove the authorization method from the header value
    _possible_token = _authorization_header.replace('Bearer', '').strip()
    # Try to parse the token into a UUID, since the tokens created by the authorization service
    # are uuids
    try:
        uuid.UUID(_possible_token)
    except ValueError:
        _logger.warning('%s:%s - %s - The bearer token is not in the correct format [REJECTED '
                        'REQUEST]',
                        request.client.host, request.client.port, _req_id)
    _logger.debug('%s:%s - %s - The bearer token present in the headers seems to be correctly '
                  'formatted',
                  request.client.host, request.client.port, _req_id)
    _logger.debug('%s:%s - %s - Requesting Token Introspection from the authorization service',
                  request.client.host, request.client.port, _req_id)
    # Create a new token introspection request
    _introspection_request = TokenIntrospectionRequest(
        bearer_token=_possible_token
    )
    # Transmit the request
    _id = _amqp_client.send(_introspection_request.json(by_alias=True), _amqp_exchange)
    _logger.debug('%s:%s - %s - Waiting for response from the authorization service',
                  request.client.host, request.client.port, _req_id)
    # Wait for a maximum of 10s for the response
    # Try getting the response
    _raw_introspection_response = _amqp_client.await_response(_id, timeout=10.0)
    if _raw_introspection_response is None:
        _logger.debug('%s:%s - %s - The authorization service did not respond in time',
                      request.client.host, request.client.port, _req_id)
        return JSONResponse(
            status_code=512,
            content={
                "error": "token_introspection_timeout"
            },
            headers={
                'Retry-After': '10'
            }
        )
    _logger.debug('%s:%s - %s - Received a response from the authorization service',
                  request.client.host, request.client.port, _req_id)
    # Parse the bytes to a dict
    _introspection_response: dict = ujson.loads(_raw_introspection_response)
    if 'active' not in _introspection_response:
        _logger.warning('%s:%s - %s - The authorization service did not respond in the correct '
                        'way', request.client.host, request.client.port, _req_id)
        return JSONResponse(
            status_code=512,
            content={
                'error': 'token_introspection_failed'
            }
        )
    if not _introspection_response['active']:
        _logger.warning('%s:%s - %s - The authorization service rejected the token',
                        request.client.host, request.client.port, _req_id)
        if _introspection_response['reason'] == 'token_error':
            _logger.warning('%s:%s - %s - The token is invalid (either never created or expired)',
                            request.client.host, request.client.port, _req_id)
            return JSONResponse(
                status_code=401,
                content={
                    "error": "invalid_token"
                }
            )
        elif _introspection_response['reason'] == 'insufficient_scope':
            _logger.warning('%s:%s - %s - The user has no privileges to access this resource',
                            request.client.host, request.client.port, _req_id)
            return JSONResponse(
                status_code=403,
                content={
                    "error": "insufficient_scope"
                }
            )
    return await call_next(request)


@geo_data_rest.get(
    path='/geo_operations/within'
)
async def geo_operations(
        layer_name: str = Query(default=...),
        layer_resolution: str = Query(default=...),
        object_names: list[str] = Query(default=...)
):
    """Get the GeoJson and names of the Objects which are within the specified layer resolution
    and the selected object(s)

    :param layer_name:
    :param layer_resolution:
    :param object_names:
    :return:
    """
    print(layer_name, layer_resolution, object_names)
    # Get the configuration of the requested layer, if it exits
    config = _map_layers.get(layer_name)
    if config is None:
        raise HTTPException(status_code=404, detail='Layer not configured')
    # Get the resolution
    resolution = None
    for res in config.resolutions:
        if res.name == layer_resolution:
            resolution = res
            break
    if resolution is None:
        raise HTTPException(status_code=404, detail='Resolution not found')
    configs = []
    for res in config.resolutions:
        if res.name in resolution.contains:
            configs.append(res)
    _connection = database.engine().connect()
    _contained_objects = {}
    for conf in configs:
        for object_name in object_names:
            _raw_query = "SELECT name, st_asgeojson(geom) " \
                         "FROM {} " \
                         "WHERE st_within(geom, ( " \
                         "SELECT geom FROM {} WHERE {}.name = '{}')) " \
                         "ORDER BY name"
            _query = _raw_query.format(
                conf.table_name, resolution.table_name, resolution.table_name, object_name
            )
            results = _connection.execute(_query).fetchall()
            _objects = {}
            for name, geojson in results:
                _objects.update({name: ujson.loads(geojson)})
            _contained_objects.update({conf.name: _objects})
    return _contained_objects


@geo_data_rest.get(
    path='/{layer_name}/{layer_resolution}',
    status_code=200
)
async def get_layer(
        layer_name: str = Path(default=..., title='Name of the Layer'),
        layer_resolution: str = Path(default=..., title='Resolution of the Layer')
):
    # Get the configuration of the requested layer, if it exits
    config = _map_layers.get(layer_name)
    if config is None:
        raise HTTPException(status_code=404, detail='Layer not configured')
    # Get the resolution
    resolution = None
    for res in config.resolutions:
        if res.name == layer_resolution:
            resolution = res
            break
    if resolution is None:
        raise HTTPException(status_code=404, detail='Resolution not found')
    # Access the table and get the name and geojson contents
    _raw_query = 'SELECT name, key, st_asgeojson(geom) as geojson from {}.{}'
    _query = text(_raw_query.format(config.database_schema, resolution.table_name))
    # Connect to the database
    _connection = database.engine().connect()
    result = _connection.execute(_query).fetchall()
    _object_list: list[dict[str, Union[str, dict]]] = []
    for name, key, geojson in result:
        _object = {
            'name':    name,
            'key': key,
            'geojson': ujson.loads(geojson)
        }
        _object_list.append(_object)
    return _object_list
