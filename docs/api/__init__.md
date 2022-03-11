---
sidebar_label: api
title: api
---

FastAPI Implementation of the REST-Service for receiving GeoJSONs


#### service\_startup

```python
@geo_data_rest.on_event('startup')
async def service_startup()
```

Handle the service startup


#### handle\_shutdown

```python
@geo_data_rest.on_event('shutdown')
async def handle_shutdown()
```

Handle the service shutdown


#### check\_user\_scope

```python
@geo_data_rest.middleware('http')
async def check_user_scope(request: Request, call_next)
```

This middleware will validate the authorization token present in the incoming request for

the scope that is assigned to it. This validation will be done via AMQP

**Arguments**:

- `request`: The incoming request
- `call_next`: The next thing that should happen

**Returns**:

The response

#### geo\_operations

```python
@geo_data_rest.get(
    path='/geo_operations/within'
)
async def geo_operations(layer_name: str = Query(default=...), layer_resolution: str = Query(default=...), object_names: list[str] = Query(default=...))
```

Get the GeoJson and names of the Objects which are within the specified layer resolution

and the selected object(s)

**Arguments**:

- `layer_name`: 
- `layer_resolution`: 
- `object_names`: 

