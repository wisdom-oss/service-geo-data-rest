---
sidebar_label: security
title: api.security
---

security related functions for the rest api


#### request\_is\_authorized

```python
def request_is_authorized(request: Request, request_id: str) -> str
```

Check if the headers contain a valid bearer token

**Arguments**:

- `request`: The request
- `request_id`: The request id which is associated with this request

**Returns**:

True if the token is valid

