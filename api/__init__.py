"""Package containing the code which will be the API later on"""
import datetime
import email.utils
import hashlib
import http
import logging
import typing

import amqp_rpc_client
import fastapi
import py_eureka_client.eureka_client
import pytz as pytz
import sqlalchemy.exc
import ujson
import yaml

import api.handler
import configuration
import database
import exceptions
import models.internal
import tools
from api import security

# %% Global Clients
_amqp_client: typing.Optional[amqp_rpc_client.Client] = None
_service_registry_client: typing.Optional[py_eureka_client.eureka_client.EurekaClient] = None

# %% API Setup
service = fastapi.FastAPI()
service.add_exception_handler(exceptions.APIException, api.handler.handle_api_error)
service.add_exception_handler(fastapi.exceptions.RequestValidationError, api.handler.handle_request_validation_error)
service.add_exception_handler(sqlalchemy.exc.IntegrityError, api.handler.handle_integrity_error)

# %% Configurations
_security_configuration = configuration.SecurityConfiguration()

# %% Preparation for layer requests
_layer_configs = yaml.safe_load(open("./configuration/layers.yaml"))
layers: typing.Dict[str, models.internal.LayerConfiguration] = {}
for _layer_config in _layer_configs:
    layers.update({_layer_config["name"]: models.internal.LayerConfiguration.parse_obj(_layer_config)})

# %% Middlewares
@service.middleware("http")
async def etag_comparison(request: fastapi.Request, call_next):
    """
    A middleware which will hash the request path and all parameters transferred to this
    microservice and will check if the hash matches the one of the ETag which was sent to the
    microservice. Furthermore, it will take the generated hash and append it to the response to
    allow caching

    :param request: The incoming request
    :type request: fastapi.Request
    :param call_next: The next call after this middleware
    :type call_next: callable
    :return: The result of the next call after this middle ware
    :rtype: fastapi.Response
    """
    # Access all parameters used for creating the hash
    path = request.url.path
    query_parameter = dict(request.query_params)
    content_type = request.headers.get("Content-Type", "text/plain")

    if content_type == "application/json":
        try:
            body = ujson.loads(await request.body())
        except ValueError as error:
            body = (await request.body()).decode("utf-8")
    else:
        body = (await request.body()).decode("utf-8")

    # Now iterate through all query parameters and make sure they are sorted if they are lists
    for key, value in dict(query_parameter).items():
        # Now check if the value is a list
        if isinstance(value, list):
            query_parameter[key] = sorted(value)

    query_dict = {
        "request_path": path,
        "request_query_parameter": query_parameter,
        "request_body": body,
    }
    query_data = ujson.dumps(query_dict, ensure_ascii=False, sort_keys=True)
    # Now create a hashsum of the query data
    query_hash = hashlib.sha3_256(query_data.encode("utf-8")).hexdigest()
    # Now access the headers of the request and check for the If-None-Match Header
    if_none_match_value = request.headers.get("If-None-Match")
    if_modified_since_value = request.headers.get("If-Modified-Since")
    if if_modified_since_value is None:
        if_modified_since_value = datetime.datetime.fromtimestamp(0, tz=pytz.UTC)
    else:
        if_modified_since_value = email.utils.parsedate_to_datetime(if_modified_since_value)
    # Get the last update of the schema from which the service gets its data from
    # TODO: Set your schema name here
    last_database_modification = tools.get_last_schema_update("geodata", database.engine)
    data_changed = if_modified_since_value < last_database_modification
    if query_hash == if_none_match_value and not data_changed:
        return fastapi.Response(status_code=304, headers={"ETag": f"{query_hash}"})
    else:
        response: fastapi.Response = await call_next(request)
        response.headers.append("ETag", f"{query_hash}")
        response.headers.append("Last-Modified", email.utils.format_datetime(last_database_modification))
        return response


# %% Routes
@service.get("/{layer_name}/{layer_resolution}")
async def scoped_hello(
    layer_name: str = fastapi.Path(default=..., title="Name of the Layer"),
    layer_resolution: str = fastapi.Path(default=..., title="Resolution of the Layer"),
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    # Try to pull the configuration of the specified layer
    layer_config = layers.get(layer_name, None)
    if layer_config is None:
        raise exceptions.APIException(
            error_code="LAYER_NOT_FOUND",
            error_title="Layer not found",
            error_description="The requested layer has not been configured",
            http_status=http.HTTPStatus.NOT_FOUND,
        )
    resolution = [res for res in layer_config.resolutions if res.name == layer_resolution]
    if len(resolution) == 0:
        raise exceptions.APIException(
            error_code="RESOLUTION_NOT_FOUND",
            error_title="Layer resolution not found",
            error_description="The requested resolution of the layer has not been configured",
            http_status=http.HTTPStatus.NOT_FOUND,
        )
    _geodata_query = "SELECT name, key, st_asgeojson(geom) from {}.{}".format(
        layer_config.database_schema, resolution[0].table_name
    )
    query_result = database.engine.execute(sqlalchemy.text(_geodata_query)).all()
    objects = []
    for name, key, geojson in query_result:
        objects.append({"name": name, "key": key, "geojson": ujson.loads(geojson)})
    return objects
