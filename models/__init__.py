"""Module containing all data models"""
from pydantic import BaseModel as PydanticBaseModel


class BaseModel(PydanticBaseModel):
    """A new base model used in this service with some preset configuration values"""
    
    class Config:
        """Configuration of the new BaseModel"""
        
        allow_population_by_field_name = True
        orm_mode = True
