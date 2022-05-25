---
sidebar_label: security
title: api.security
---

#### is\_authorized\_user

```python
def is_authorized_user(
    scopes: fastapi.security.SecurityScopes,
    access_token: str = fastapi.Depends(__wisdom_central_auth)
) -> typing.Union[bool, models.internal.UserAccount]
```

Check if the user calling this service is authorized.

This security dependency needs to be used as fast api dependency in the methods

**Arguments**:

- `scopes` (`list`): The scopes this used needs to have to access this service
- `access_token` (`str`): The access token used by the user to access the service

**Raises**:

- `exceptions.APIException`: The user is not authorized to access this service

**Returns**:

`bool`: Status of the authorization

