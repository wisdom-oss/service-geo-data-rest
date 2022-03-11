"""Module containing models related to the geo data"""
from typing import List

from pydantic import Field

from .. import BaseModel


class LayerResolution(BaseModel):
    """Model describing how a layer resolution is created"""
    
    name: str = Field(
        default=...,
        alias='name'
    )
    """
    Resolution Name
    
    The name of the resolution which may be used in requests to specify the resolution
    """
    
    table_name: str = Field(
        default=...,
        alias='table'
    )
    """
    Table Name
    
    The name of the table in which the geospatial data for this resolution is stored
    """
    
    is_default: bool = Field(
        default=False,
        alias='default'
    )
    """
    Default Resolution Indicator
    
    If this is set to true, use this resolution for this layer if the request did not specify a
    resolution. [default: ``False``]
    """
    

class LayerConfiguration(BaseModel):
    """The configuration of a layer"""
    
    database_schema: str = Field(
        default=...,
        alias='db_schema'
    )
    """
    Database Schema
    
    The name of the schema in which the tables are stored containing the geospatial data
    """
    
    resolutions: List[LayerResolution] = Field(
        default=...,
        alias='resolutions'
    )
    """
    Layer Resolutions
    
    The different resolutions this layer has to offer
    """
