"""FastAPI Implementation of the REST-Service for receiving GeoJSONs"""
import logging
import time

from fastapi import FastAPI as RESTfulAPI
from fastapi import Request

GEO_DATA_REST = RESTfulAPI()

_logger = logging.getLogger('REST-API')


@GEO_DATA_REST.middleware('http')
async def check_user_scope(request: Request, next_call):
    """This middleware will validate the authorization token present in the incoming request for
    the scope that is assigned to it. This validation will be done via AMQP
    
    :param request: The incoming request
    :param next_call: The next thing that should happen
    :return: The response
    """
    # TODO: Implement check
    response = await next_call(request)
    return response

