---
sidebar_label: internal
title: models.internal
---

## ServiceScope Objects

```python
class ServiceScope(__BaseModel)
```

#### name

The name of the scope this service uses


#### description

The description of the scope this service uses


#### value

The string by which this scope is identifiable in a OAuth 2.0 scope string


## UserAccount Objects

```python
class UserAccount(__BaseModel)
```

#### id

Internal Account ID


#### first\_name

The first name of the user who is the owner of the account


#### last\_name

The last name of the user who is the owner of the account


#### username

The username of the account


## LayerResolution Objects

```python
class LayerResolution(__BaseModel)
```

Model describing how a layer resolution is created


#### name

Resolution Name

The name of the resolution which may be used in requests to specify the resolution


#### table\_name

Table Name

The name of the table in which the geospatial data for this resolution is stored


#### is\_default

Default Resolution Indicator

If this is set to true, use this resolution for this layer if the request did not specify a
resolution. [default: ``False``]


#### contains

Contained resolutions

The names of the resolutions which are more granular than this one.


## LayerConfiguration Objects

```python
class LayerConfiguration(__BaseModel)
```

The configuration of a layer


#### database\_schema

Database Schema

The name of the schema in which the tables are stored containing the geospatial data


#### resolutions

Layer Resolutions

The different resolutions this layer has to offer


