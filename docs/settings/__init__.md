---
sidebar_label: settings
title: settings
---

Settings which change the behaviour of the service


## ServiceSettings Objects

```python
class ServiceSettings(BaseSettings)
```

Settings which are directly related to the service


#### name

Application name

The name of the service which will be used for registering at the service registry and used as
part of the authorization flow at the authorization service for various operations


#### http\_port

HTTP Port

The HTTP port which will be bound by uvicorn at the startup for allowing HTTP access to this
service


#### log\_level

Logging Level

The level of logging which the root logger will be configured to


## Config Objects

```python
class Config()
```

Configuration of the service settings


#### env\_file

Allow loading these values from the specified file


## ServiceRegistrySettings Objects

```python
class ServiceRegistrySettings(BaseSettings)
```

Settings related to the connection to the service registry


#### host

Service registry host (required)

The hostname or ip address of the service registry on which this service shall register itself


#### port

Service registry port

The port on which the service registry listens on, defaults to 8761


## Config Objects

```python
class Config()
```

Configuration of the service registry settings


#### env\_file

The location of the environment file from which these values may be loaded


## AMQPSettings Objects

```python
class AMQPSettings(BaseSettings)
```

Settings related to AMQP-based communication


#### dsn

AMQP Address

The address pointing to a AMQP-enabled message broker which shall be used for internal
communication between the services


#### auth\_exchange

AMQP Authorization Service Exchange

The exchange which is bound by the authorization service, defaults to `authorization-service`


## Config Objects

```python
class Config()
```

Configuration of the AMQP related settings


#### env\_file

The file from which the settings may be read


## DatabaseSettings Objects

```python
class DatabaseSettings(BaseSettings)
```

Settings related to the connections to the geo-data server


#### dsn

PostgreSQL Database Service Name

An URI pointing to the installation of a Postgres database which has the PostGIS extensions
installed and activated


## Config Objects

```python
class Config()
```

Configuration of the AMQP related settings


#### env\_file

The file from which the settings may be read


