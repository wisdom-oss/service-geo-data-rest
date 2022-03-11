---
sidebar_label: geo
title: models.geo
---

Module containing models related to the geo data


## LayerResolution Objects

```python
class LayerResolution(BaseModel)
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
class LayerConfiguration(BaseModel)
```

The configuration of a layer


#### database\_schema

Database Schema

The name of the schema in which the tables are stored containing the geospatial data


#### resolutions

Layer Resolutions

The different resolutions this layer has to offer


