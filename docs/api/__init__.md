---
sidebar_label: api
title: api
---

Package containing the code which will be the API later on


#### etag\_comparison

```python
@service.middleware("http")
async def etag_comparison(request: fastapi.Request, call_next)
```

A middleware which will hash the request path and all parameters transferred to this

microservice and will check if the hash matches the one of the ETag which was sent to the
microservice. Furthermore, it will take the generated hash and append it to the response to
allow caching

**Arguments**:

- `request` (`fastapi.Request`): The incoming request
- `call_next` (`callable`): The next call after this middleware

**Returns**:

`fastapi.Response`: The result of the next call after this middle ware

