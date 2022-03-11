"""Settings which change the behaviour of the service"""
from pydantic import BaseSettings, Field, AmqpDsn, SecretStr, PostgresDsn


class ServiceSettings(BaseSettings):
    """
    Settings which are directly related to the service
    """

    name: str = Field(
        default='geo-data-rest',
        title='Service Name',
        description='The name of the service which is used for registering at the service '
                    'registry and authorization at the authorization service for various '
                    'operations',
        env='SERVICE_NAME'
    )
    """
    Application name
    
    The name of the service which will be used for registering at the service registry and used as
    part of the authorization flow at the authorization service for various operations
    """
    
    http_port: int = Field(
        default=5000,
        title='HTTP Port',
        description='The http port which will be bound at startup for the HTTP Access to this '
                    'service',
        env='SERVICE_HTTP_PORT'
    )
    """
    HTTP Port
    
    The HTTP port which will be bound by uvicorn at the startup for allowing HTTP access to this
    service
    """
    
    log_level: str = Field(
        default='INFO',
        title='Logging Level',
        description='The level of logging which the root logger will use',
        env='LOG_LEVEL'
    )
    """
    Logging Level
    
    The level of logging which the root logger will be configured to
    """
    
    class Config:
        """Configuration of the service settings"""
        
        env_file = '.application.env'
        """Allow loading these values from the specified file"""


class ServiceRegistrySettings(BaseSettings):
    """Settings related to the connection to the service registry"""
    
    host: str = Field(
        default=...,
        title='Service registry host',
        description='The hostname or ip address of the service registry on which this service '
                    'shall register itself',
        env='SERVICE_REGISTRY_HOST'
    )
    """
    Service registry host (required)
    
    The hostname or ip address of the service registry on which this service shall register itself
    """
    
    port: int = Field(
        default=8761,
        title='Service registry port',
        description='The port on which the service registry listens on, defaults to 8761',
        env='SERVICE_REGISTRY_PORT'
    )
    """
    Service registry port
    
    The port on which the service registry listens on, defaults to 8761
    """
    
    class Config:
        """Configuration of the service registry settings"""
        
        env_file = '.registry.env'
        """The location of the environment file from which these values may be loaded"""


class AMQPSettings(BaseSettings):
    """Settings related to AMQP-based communication"""
    
    dsn: AmqpDsn = Field(
        default=...,
        title='AMQP Address',
        description='The amqp address containing the login information into a message broker',
        env='AMQP_DSN'
    )
    """
    AMQP Address
    
    The address pointing to a AMQP-enabled message broker which shall be used for internal
    communication between the services
    """
    
    auth_exchange: str = Field(
        default='authorization-service',
        title='Name of the exchange',
        description='Name of the exchange which is bound by the authorization service',
        env='AMQP_AUTH_EXCHANGE'
    )
    """
    AMQP Authorization Service Exchange
    
    The exchange which is bound by the authorization service, defaults to `authorization-service`
    """
    
    class Config:
        """Configuration of the AMQP related settings"""
        
        env_file = '.amqp.env'
        """The file from which the settings may be read"""
        

class DatabaseSettings(BaseSettings):
    """Settings related to the connections to the geo-data server"""
    
    dsn: PostgresDsn = Field(
        default=...,
        title='PostgreSQL Database Service Name',
        description='A uri pointing to the postgres database containing the geo data',
        env='DATABASE_DSN'
    )
    """
    PostgreSQL Database Service Name
    
    An URI pointing to the installation of a Postgres database which has the PostGIS extensions
    installed and activated
    """

    class Config:
        """Configuration of the AMQP related settings"""
    
        env_file = '.database.env'
        """The file from which the settings may be read"""
