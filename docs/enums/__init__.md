---
sidebar_label: enums
title: enums
---

## AMQPAction Objects

```python
class AMQPAction(str, enum.Enum)
```

The actions which are available for this service


#### CHECK\_TOKEN\_SCOPE

Check the scope of a token and return if the token is valid and has the scope


#### ADD\_SCOPE

Add a scope to the authorization system


#### EDIT\_SCOPE

Edit a scope already in the authorization system


#### CHECK\_SCOPE

Check if a scope is already present in the system


## TokenIntrospectionFailure Objects

```python
class TokenIntrospectionFailure(str, enum.Enum)
```

The reasons why a token introspection has failed and did not return that the token is valid


#### INVALID\_TOKEN

The token either has an invalid format or was not found in the database


#### EXPIRED

The tokens TTL as expired


#### TOKEN\_USED\_TOO\_EARLY

The token has been used before it&#x27;s creation time


#### NO\_USER\_ASSOCIATED

The token has no user associated to it


#### USER\_DISABLED

The user associated to the account has been disabled


#### MISSING\_PRIVILEGES

The scopes associated to this token are not matching the one required to access this endpoint


## Resolution Objects

```python
class Resolution(str, enum.Enum)
```

The resolutions by name mapped to the length of the keys


