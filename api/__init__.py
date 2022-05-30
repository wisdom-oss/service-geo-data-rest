"""Package containing the code which will be the API later on"""
import datetime
import email.utils
import hashlib
import http
import typing

import amqp_rpc_client
import fastapi
import geoalchemy2.functions
import py_eureka_client.eureka_client
import pytz as pytz
import sqlalchemy.exc
import sqlalchemy.dialects
import starlette.middleware.gzip
import ujson

import api.handler
import configuration
import database
import database.tables
import enums
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
service.add_middleware(starlette.middleware.gzip.GZipMiddleware, minimum_size=0)

# %% Configurations
_security_configuration = configuration.SecurityConfiguration()
if _security_configuration.scope_string_value is None:
    scope = models.internal.ServiceScope.parse_file("./configuration/scope.json")
    _security_configuration.scope_string_value = scope.value

# %% Custom Mappings
key_length_mapping = {
    enums.Resolution.state: 2,
    enums.Resolution.district: 5,
    enums.Resolution.administration: 9,
    enums.Resolution.municipal: 12,
}


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
@service.get("/")
async def scoped_hello(
    resolution: typing.Optional[enums.Resolution] = fastapi.Query(default=None),
    keys: typing.Optional[list[str]] = fastapi.Query(default=None, alias="key", regex=r"^\d{1,12}$"),
    user: typing.Union[models.internal.UserAccount, bool] = fastapi.Security(
        security.is_authorized_user, scopes=[_security_configuration.scope_string_value]
    ),
):
    if resolution is None and keys is None:
        raise exceptions.APIException(
            error_code="INVALID_QUERY",
            error_title="Invalid Parameter Combination",
            error_description="Your need to set at least one query parameter for a successful request",
            http_status=http.HTTPStatus.BAD_REQUEST,
        )
    # Build the regex string
    if keys is None:
        shape_query = sqlalchemy.select(
            [
                database.tables.shapes.c.name,
                database.tables.shapes.c.key,
                database.tables.shapes.c.nuts_key,
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(
                        geoalchemy2.functions.ST_Transform(database.tables.shapes.c.geom, 4326)
                    ),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson"),
            ],
            sqlalchemy.func.length(database.tables.shapes.c.key) == key_length_mapping.get(resolution),
        )
        box_query = sqlalchemy.select(
            [
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(geoalchemy2.functions.ST_Extent(database.tables.shapes.c.geom)),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson")
            ],
            sqlalchemy.func.length(database.tables.shapes.c.key) == key_length_mapping.get(resolution),
        )
    elif resolution is None:
        shape_query = sqlalchemy.select(
            [
                database.tables.shapes.c.name,
                database.tables.shapes.c.key,
                database.tables.shapes.c.nuts_key,
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(
                        geoalchemy2.functions.ST_Transform(database.tables.shapes.c.geom, 4326)
                    ),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson"),
            ],
            database.tables.shapes.c.key.in_(keys),
        )
        box_query = sqlalchemy.select(
            [
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(geoalchemy2.functions.ST_Extent(database.tables.shapes.c.geom)),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson")
            ],
            database.tables.shapes.c.key.in_(keys),
        )
    else:
        regex = r""
        for key in keys:
            if len(key) < key_length_mapping.get(resolution):
                regex += rf"^{key}\d+$|"
            else:
                regex += rf"^{key}$|"
        regex = regex.strip("|")
        shape_query = sqlalchemy.select(
            [
                database.tables.shapes.c.name,
                database.tables.shapes.c.key,
                database.tables.shapes.c.nuts_key,
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(
                        geoalchemy2.functions.ST_Transform(database.tables.shapes.c.geom, 4326)
                    ),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson"),
            ],
            sqlalchemy.and_(
                database.tables.shapes.c.key.regexp_match(regex),
                sqlalchemy.func.length(database.tables.shapes.c.key) == key_length_mapping.get(resolution),
            ),
        )
        box_query = sqlalchemy.select(
            [
                sqlalchemy.cast(
                    geoalchemy2.functions.ST_AsGeoJSON(geoalchemy2.functions.ST_Extent(database.tables.shapes.c.geom)),
                    sqlalchemy.dialects.postgresql.JSONB,
                ).label("geojson")
            ],
            sqlalchemy.and_(
                database.tables.shapes.c.key.regexp_match(regex),
                sqlalchemy.func.length(database.tables.shapes.c.key) == key_length_mapping.get(resolution),
            ),
        )
    shape_query_result = database.engine.execute(shape_query).all()
    box_query_result = database.engine.execute(box_query).first()
    if len(shape_query_result) == 0:
        return fastapi.Response(status_code=204)
    return {"box": box_query_result["geojson"]["coordinates"][0], "shapes": shape_query_result}
