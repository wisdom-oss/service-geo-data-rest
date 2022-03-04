"""FastAPI Implementation of the REST-Service for receiving GeoJSONs"""
import logging
import re
from typing import Optional

import ujson as ujson
from fastapi import FastAPI as RESTfulAPI, Request
from starlette import status
from starlette.responses import JSONResponse

import amqp
from models.amqp import TokenValidationRequest

GEO_DATA_REST = RESTfulAPI()

_logger = logging.getLogger('REST-API')
_amqp_client: Optional[amqp.RPCClient] = None


@GEO_DATA_REST.middleware('http')
async def check_user_scope(request: Request, next_call):
    """This middleware will validate the authorization token present in the incoming request for
    the scope that is assigned to it. This validation will be done via AMQP
    
    :param request: The incoming request
    :param next_call: The next thing that should happen
    :return: The response
    """
    # Access the request headers
    headers = request.headers
    # Check if the headers contain an Authorization header
    authorization_header_present = True if headers.get('Authorization', None) is not None else False
    if not authorization_header_present:
        # Since there was no authorization header present return an error response
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            headers={
                'WWW-Authenticate': 'Bearer'
            },
            content={
                'error': 'no_auth_information_present'
            }
        )
    # Since a header was found, try to extract the token if a token is present via regex
    regex = r"[Bb]earer ([0-9a-fA-F]{8}\b-(?:[0-9a-fA-F]{4}\b-){3}[0-9a-fA-F]{12})"
    header_value = headers.get('Authorization')
    if match := re.match(regex, header_value):
        # Get the token
        token = match.group(1)
        # Create a new validation request
        _validation_request = TokenValidationRequest(
            bearer_token=token
        )
        _id = _amqp_client.send_message(_validation_request.json(by_alias=True))
        # Wait for the response to be received
        if _amqp_client.message_events[_id].wait():
            # Load the response
            _validation_response: dict = ujson.loads(_amqp_client.responses[_id])
            if ('active' in _validation_response) and _validation_response['active'] is True:
                response = await next_call(request)
                return response
            else:
                return JSONResponse(
                    status_code=status.HTTP_403_FORBIDDEN,
                    content={
                        'error': 'missing_permissions',
                        'description': 'The authorized user has no permissions to access this '
                                       'resource'
                    }
                )
    else:
        return JSONResponse(
            status_code=status.HTTP_400_BAD_REQUEST,
            headers={
                'WWW-Authenticate': 'Bearer'
            },
            content={
                'error': 'no_auth_information_present'
            }
        )

