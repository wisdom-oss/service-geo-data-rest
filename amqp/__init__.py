"""An asynchronous RPC Client which sends and receives messages from a target service"""
import logging
import secrets
import threading
import time
import typing
import uuid

import pika
from pika.adapters.blocking_connection import BlockingChannel
from pika.spec import Basic, BasicProperties

from settings import AMQPSettings, ServiceSettings


class RPCClient:
    """An asynchronous AMQP RPC Client working with events"""
    
    __internal_messaging_lock = threading.Lock()
    responses: typing.Dict[str, str] = {}
    message_events: typing.Dict[str, threading.Event] = {}
    
    def __init__(
            self
    ):
        """
        Create a new RPC Client
        
        To successfully create a new RPC Client the following environment variables need to be set:
          - AMQP_DSN
          
        During the creation the environment variable will be used to determine the connection
        parameters. Refer to the documentation for more information
        """
        # Read the AMQP Settings again
        self.__settings = AMQPSettings()
        self.__service_settings = ServiceSettings()
        self.__logger = logging.getLogger('AMQP-RPC')
        self.__logger.info('Creating a new AMQP RPC Client')
        # Create a unique name for this connection
        self.name = self.__service_settings.name + '-' + str(secrets.token_hex(nbytes=4))
        # Read the amqp dsn and create connection parameters from it
        _amqp_params = pika.URLParameters(self.__settings.dsn)
        # Set some additional client properties
        _amqp_params.client_properties = {
            'connection_name': self.name,
            'service_name':    self.__service_settings.name
        }
        # Disable pikas logging
        logging.getLogger("pika").setLevel(logging.WARNING)
        # Create a new Blocking Connection
        self.__connection = pika.BlockingConnection(_amqp_params)
        # Open a new channel to the message broker
        self.__channel = self.__connection.channel()
        # Create a new queue which is exclusive and will be deleted when the client disconnects
        self.__queue = self.__channel.queue_declare('', passive=False, durable=False,
                                                    exclusive=True, auto_delete=True)
        # Save the name of the queue to use later on as reply-to property of messages
        self.__callback_queue = self.__queue.method.queue
        self.__receive_messages = True
        # Create a thread for processing data events
        self.__processing_thread = threading.Thread(
            target=self.__process_new_messages,
            daemon=True
        )
        self.__processing_thread.start()
    
    def __process_new_messages(self):
        """Process new messages while consuming messages"""
        self.__consumer = self.__channel.basic_consume(
            self.__callback_queue,
            self.__new_message_received,
            auto_ack=False,
            exclusive=True
        )
        
        while True:
            if not self.__receive_messages:
                break
            with self.__internal_messaging_lock:
                self.__connection.process_data_events()
                time.sleep(.05)
        
        self.__channel.basic_cancel(self.__consumer)
        
    def __new_message_received(
            self,
            channel: BlockingChannel,
            method: Basic.Deliver,
            properties: BasicProperties,
            content: bytes
    ):
        """Handle a new message by adding it to the stack of responses and calling the related event
        
        :param channel: The channel on which the message was received
        :param method: Information about the delivery
        :param properties: Properties of the incoming message
        :param content: The content of the message
        """
        self.responses[properties.correlation_id] = content.decode('utf-8')
        self.__channel.basic_ack(method.delivery_tag)
        self.message_events[properties.correlation_id].set()
        
    def send_message(self, content: str) -> str:
        """Send the message content to the exchange specified in the settings
        
        :param content: The content of the message
        :return: The message uuid which can be used to await and get the message
        """
        _message_id = str(uuid.uuid1())
        # Create a new event for the to be awaited response
        self.message_events[_message_id] = threading.Event()
        with self.__internal_messaging_lock:
            # Send the message to the message broker
            self.__channel.basic_publish(
                exchange=self.__settings.auth_exchange,
                routing_key='',
                properties=pika.BasicProperties(
                    reply_to=self.__callback_queue,
                    correlation_id=_message_id,
                    content_encoding='utf-8'
                ),
                body=content.encode('utf-8')
            )
        return _message_id
        
        