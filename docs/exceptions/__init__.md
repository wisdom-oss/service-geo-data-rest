---
sidebar_label: exceptions
title: exceptions
---

## APIException Objects

```python
class APIException(Exception)
```

An exception for returning any error that happened in the API


#### \_\_init\_\_

```python
def __init__(
    error_code: str,
    error_title: typing.Optional[str] = None,
    error_description: typing.Optional[str] = None,
    http_status: typing.Union[http.HTTPStatus,
                              int] = http.HTTPStatus.INTERNAL_SERVER_ERROR)
```

Create a new API exception

**Arguments**:

- `error_code` (`str`): The error code of the exception
- `error_title` (`str`): The title of the exception
- `error_description` (`str`): The description of the exceptions
- `http_status` (`http.HTTPStatus`): The HTTP Status that will be sent back

