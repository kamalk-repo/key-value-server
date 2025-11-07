# key-value-server

Implementation of Key-Value server project in course CS-744, Fall, 2025, IIT Bombay.

# Overview

This project implements a custom key–value storage server using the Go programming language with a focus on performance benchmarking and bottleneck analysis.

The server provides a lightweight, REST-based interface for performing Create, Read, Update, and Delete (CRUD) operations on key–value pairs.

It features an integrated caching subsystem supporting both write-back and write-through policies, enabling analysis of latency, throughput, and consistency trade-offs in different caching strategies.

All persistent data is stored in a MySQL database, providing a durable backend for evaluating database–cache interactions and write propagation performance.

The main objective of this project is to study the performance and identify bottlenecks under various workloads.


# API Routes

#### Read Key

```http
  GET /kv_store?key=keyId
```

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `key` | `int` | Id of the key to fetch |

```

Sample Response:
  {
    "Status": 0,
    "Error": "No error",
    "Data": {
      "Key": 3,
      "Value": "KeyValue"
    }
  }
```

#### Add key

```http
  POST /kv_store
  Content-type: application/json
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `Key`      | `int` | Id of the key to store |
| `Value`      | `string` | Value of the key to store |

```
Sample request:
  {
    "Key":3,
    "Value":"KeyValue"
  }

Sample Response:
  {
    "Status": 0,
    "Error": "No error",
    "Data": "Inserted key: 3"
  }

```

#### Update key

```http
  PUT /kv_store
  Content-type: application/json
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `Key`      | `int` | Id of the key to update |
| `Value`      | `string` | Value of the key to update |

```
Sample request:
  {
    "Key":3,
    "Value":"KeyValue"
  }
  
Sample Response:
  {
    "Status": 0,
    "Error": "No error",
    "Data": "Updated key: 3"
  }

```

#### Delete key

```http
  DELETE /kv_store?key=keyId
```

| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `key`      | `int` | Id of the key to delete |

```
Sample Response:
  {
    "Status": 0,
    "Error": "No error",
    "Data": "Deleted key: 3"
  }

```

# Authors

- [@Kamal](https://www.github.com/kamalk-repo)