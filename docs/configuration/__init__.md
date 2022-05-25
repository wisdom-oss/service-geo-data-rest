---
sidebar_label: configuration
title: configuration
---

## ServiceConfiguration Objects

```python
class ServiceConfiguration(pydantic.BaseSettings)
```

#### name

Microservice Name

The name of the microservice. The name will be used to identify this service and it&#x27;s instances at the service
registry


#### http\_port

HTTP Port

The http port which will be bound by the service in the container


#### logging\_level

Logging Level

The level which is used for the root logger. The root logger will display messages from this level and levels
above this one.


## AMQPConfiguration Objects

```python
class AMQPConfiguration(pydantic.BaseSettings)
```

#### dsn

AMQP Data Source Name

The data source name pointing to an installation of the RabbitMQ message broker


#### exchange

AMQP Send Exchange

The exchange to which this service will send the messages


#### authorization\_exchange

AMQP Authorization Service

The exchange to which this service will send the messages related to authorizing users and requests


## SecurityConfiguration Objects

```python
class SecurityConfiguration(pydantic.BaseSettings)
```

#### scope\_string\_value

Required Scope String Value

The scope string value of the scope which is required to access this service. If no value is set the access to
the services routes are unprotected


## ServiceRegistryConfiguration Objects

```python
class ServiceRegistryConfiguration(pydantic.BaseSettings)
```

Settings which will influence the connection to the service registry


#### host

Eureka Service Registry Host

The host on which the eureka service registry is running on.


#### port

Eureka Service Registry Port

The port on which the eureka service registry is running on.


## DatabaseConfiguration Objects

```python
class DatabaseConfiguration(pydantic.BaseSettings)
```

Settings which are related to the database connection


#### dsn

PostgreSQL data source name

The data source name (expressed as URI) pointing to the installation of the used postgresql database


