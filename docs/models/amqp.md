---
sidebar_label: amqp
title: models.amqp
---

## TokenIntrospectionRequest Objects

```python
class TokenIntrospectionRequest(_BaseModel)
```

The data model describing how a token introspection request will look like


#### action

The action that shall be executed on the authorization server


#### bearer\_token

The Bearer token that has been extracted and now shall be validated


#### scope

The scope which needs to be in the tokens scope to pass the introspection


## CreateScopeRequest Objects

```python
class CreateScopeRequest(_BaseModel)
```

The data model describing how a scope creation request will look like


#### name

The name of the new scope


#### description

The description of the new scope


#### value

String which will identify the scope


## CheckScopeRequest Objects

```python
class CheckScopeRequest(_BaseModel)
```

#### value

The value of the scope that shall tested for existence


