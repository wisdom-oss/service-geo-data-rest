# WISdoM OSS - Geospatial Data REST Service

This service allows access to the already stored geospatial data in the project.

To use this service in a non default deployment you may use the following docker compose snippet
```yaml
version: '3.8'

services:
  geospatial-data-rest:
    build: https://github.com/wisdom-oss/service-geo-data-rest#stable
    image: wisdom-oss/services/geospatial-data-rest:stable
    restart: always
    deploy:
      mode: replicated
      replicas: 1 #TODO: Increase if more replicas are needed for your deployment
    depends_on:
      - postgres
      - api-gateway
    # TODO: Set the following values to those matching your deployment
    environment:
      - CONFIG_API_GATEWAY_HOST=
      - CONFIG_API_GATEWAY_PORT=
      - CONFIG_API_GATEWAY_SERVICE_PATH=
      - CONFIG_POSTGRES_HOST=
      - CONFIG_POSTGRES_USER=
      - CONFIG_POSTGRES_PASSWORD=
```

## Configuration
The service supports the following configuration parameters by environment variables:

- `CONFIG_LOGGING_LEVEL` &#8594; Set the logging verbosity [optional, default `INFO`]
- `CONFIG_API_GATEWAY_HOST` &#8594; Set the host on which the API Gateway runs on **[required]**
- `CONFIG_API_GATEWAY_PORT` &#8594; Set the port on which the API Gateway listens on **[required]**
- `CONFIG_API_GATEWAY_SERVICE_PATH` &#8594; Set the path under which the service shall be reachable. _Do not prepend the path with `/api`. Only set the last part of the desired path_ **[required]**
- `CONFIG_POSTGRES_HOST` &#8594; The host on which the database runs on containing the geospatial data
- `CONFIG_POSTGRES_USER` &#8594; The user used to access the database
- `CONFIG_POSTGRES_PASSWORD` &#8594; The password used to access the database
- `CONFIG_HTTP_LISTEN_PORT` &#8594; The port on which the built-in webserver will listen on [optional, default `8000`]
- `CONFIG_SCOPE_FILE_PATH` &#8594; The location where the scope definition file is stored inside the docker container [optional, default `/microservice/res/scope.json]