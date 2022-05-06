"""security related functions for the rest api"""
import logging

from fastapi.security import OAuth2PasswordBearer
from starlette.requests import Request

_logger = logging.getLogger('REST-API.security')


oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/auth/oauth/token")
