---
sidebar_label: tools
title: tools
---

#### is\_host\_available

```python
async def is_host_available(host: str, port: int, timeout: int = 10) -> bool
```

Check if the specified host is reachable on the specified port

**Arguments**:

- `host`: The hostname or ip address which shall be checked
- `port`: The port which shall be checked
- `timeout`: Max. duration of the check

**Returns**:

A boolean indicating the status

