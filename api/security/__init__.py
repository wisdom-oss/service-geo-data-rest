"""security related functions for the rest api"""
import logging
import uuid
from typing import Optional

from fastapi.exceptions import HTTPException
from fastapi.security import OAuth2PasswordBearer
from starlette.requests import Request

from amqp import RPCClient
from models.amqp import TokenIntrospectionRequest

_logger = logging.getLogger('REST-API.security')


oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/auth/oauth/token")


def request_is_authorized(request: Request, request_id: str) -> str:
    """Check if the headers contain a valid bearer token
    
    :param request: The request
    :param request_id: The request id which is associated with this request
    :return: True if the token is valid
    """
    
