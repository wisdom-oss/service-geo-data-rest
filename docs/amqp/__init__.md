---
sidebar_label: amqp
title: amqp
---

An asynchronous RPC Client which sends and receives messages from a target service


## RPCClient Objects

```python
class RPCClient()
```

An asynchronous AMQP RPC Client working with events


#### \_\_init\_\_

```python
def __init__()
```

Create a new RPC Client

To successfully create a new RPC Client the following environment variables need to be set:
  - AMQP_DSN

During the creation the environment variable will be used to determine the connection
parameters. Refer to the documentation for more information


#### \_\_process\_new\_messages

```python
def __process_new_messages()
```

Process new messages while consuming messages


#### \_\_new\_message\_received

```python
def __new_message_received(channel: BlockingChannel, method: Basic.Deliver, properties: BasicProperties, content: bytes)
```

Handle a new message by adding it to the stack of responses and calling the related event

**Arguments**:

- `channel`: The channel on which the message was received
- `method`: Information about the delivery
- `properties`: Properties of the incoming message
- `content`: The content of the message

#### send\_message

```python
def send_message(content: str) -> str
```

Send the message content to the exchange specified in the settings

**Arguments**:

- `content`: The content of the message

**Returns**:

The message uuid which can be used to await and get the message

