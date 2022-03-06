"""FastAPI Implementation of the REST-Service for receiving GeoJSONs"""
import logging
import uuid
from typing import Optional

import yaml
from fastapi import FastAPI, Request, Path, HTTPException, Response
from starlette.responses import JSONResponse

from amqp import RPCClient
from api import security
from models.amqp import TokenIntrospectionRequest
from models.geo import LayerConfiguration

geo_data_rest = FastAPI()

_logger = logging.getLogger('REST-API')
_map_layers: Optional[dict[str, LayerConfiguration]] = {}
_amqp_client = RPCClient()


@geo_data_rest.on_event('startup')
async def service_startup():
    """Handle the service startup"""
    # Enable the global setup of the amqp client
    global _map_layers
    # Try to read the layers.yaml
    raw_layer_config = yaml.safe_load(open('layers.yaml'))
    # Create the configurations of the layers
    for layer in raw_layer_config:
        _map_layers.update({layer['name']: LayerConfiguration.parse_obj(layer)})

    
@geo_data_rest.middleware('http')
async def check_user_scope(request, call_next):
    """This middleware will validate the authorization token present in the incoming request for
    the scope that is assigned to it. This validation will be done via AMQP
    
    :param request: The incoming request
    :param call_next: The next thing that should happen
    :return: The response
    """
    # Access the request headers
    headers = request.headers
    # Check for a present request id and use it for logging purposes
    _req_id = headers.get('X-Request-ID', str(uuid.uuid4())).replace('-', '')
    _logger.info('%s:%s - %s - Received new request for geospatial data',
                 request.client.host, request.client.port, _req_id)
    _logger.info('%s:%s - %s - Checking the request for a valid Bearer Token',
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
    _logger.info('%s:%s - %s - Found the "Authorization" header in the request',
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
    _logger.debug('%s:%s - %s - Extracted possible bearer token from header: %s',
                  request.client.host, request.client.port, _req_id, _possible_token)
    # Try to parse the token into a UUID, since the tokens created by the authorization service
    # are uuids
    try:
        uuid.UUID(_possible_token)
    except ValueError:
        _logger.warning('%s:%s - %s - The bearer token is not in the correct format [REJECTED '
                        'REQUEST]',
                        request.client.host, request.client.port, _req_id)
    _logger.info('%s:%s - %s - The bearer token present in the headers seems to be correctly '
                 'formatted',
                 request.client.host, request.client.port, _req_id)
    _logger.info('%s:%s - %s - Requesting Token Introspection from the authorization service',
                 request.client.host, request.client.port, _req_id)
    # Create a new token introspection request
    _introspection_request = TokenIntrospectionRequest(
        bearer_token=_possible_token
    )
    # Transmit the request
    _id = _amqp_client.send_message(_introspection_request.json(by_alias=True))
    _logger.info('%s:%s - %s - Waiting for response from the authorization service',
                 request.client.host, request.client.port, _req_id)
    # Wait for a maximum of 10s for the response
    if _amqp_client.message_events[_id].wait(timeout=10):
        _logger.info('%s:%s - %s - Received a response from the authorization service',
                     request.client.host, request.client.port, _req_id)
    else:
        _logger.warning('%s:%s - %s - The authorization service did not respond in time',
                        request.client.host, request.client.port, _req_id)
        return JSONResponse(
            status_code=503,
            content={
                "error": "token_introspection_timeout"
            },
            headers={
                'Retry-After': '30'
            }
        )
    return await call_next(request)
        

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
    else:
        # Try to find the correct resolution
        return config.resolutions
    