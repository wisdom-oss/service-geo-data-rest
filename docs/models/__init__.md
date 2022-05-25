---
sidebar_label: models
title: models
---

## BaseModel Objects

```python
class BaseModel(pydantic.BaseModel)
```

The base model for all other models which has some preconfigured configuration


## Config Objects

```python
class Config()
```

The configuration that all models will inherit if it bases itself on this BaseModel


#### extra

Allow extra attributes to be assigned to the model


#### allow\_population\_by\_field\_name

Allow fields to be populated by their name and alias


#### smart\_union

Check all types of a Union to prevent converting types


