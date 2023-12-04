# Kong Response Cacher Plugin

## Overview

The Kong Response Cacher Plugin is designed to cache responses from your services for requests that result in successful HTTP response codes (less than 300). This plugin uses Redis for caching and is intended to enhance the performance and efficiency of your API gateway by reducing redundant processing and network overhead.

## Requirements

- Kong Gateway version 2.x or higher.
- Go programming language environment.
- Docker for containerization.

## Features

- Caches responses of successful API requests (HTTP status < 300).
- Utilizes Redis as the caching backend.
- Configurable cache TTL (Time-To-Live) for controlling cache expiration.
- Supports custom Redis configuration including host, port, password, and database.

## Installation

### Building the Go Plugin

1. Set up your Kong environment with Go plugin support.
2. Clone or download this repository to your local machine.
3. Navigate to the project directory.
4. Build the Go plugin with the following command:
   ```bash
   go build -o response-cacher
    ```

### Building the Docker Image
In the project directory (where the Dockerfile is located), run:
```bash
docker build -t kong-response-cacher .
```

### Running the Docker Container
Run the following command to start the Kong Gateway with the Response Cacher Plugin:
  ```bash
docker run -p 8000:8000 -p 8443:8443 -p 8001:8001 -p 8444:8444 kong-response-cacher
```

### Configuration
To enable the plugin on a specific service, use the following command:

```bash

curl -X POST http://<kong_admin_api>:8001/services/<service_name>/plugins \
    --data "name=response-cacher" \
    --data "config.cache_ttl=60" \
    --data "config.redis_host=<redis_host>" \
    --data "config.redis_port=<redis_port>" \
    --data "config.redis_password=<redis_password>" \
    --data "config.redis_database=<redis_database>"

```
Replace <service_name>, <redis_host>, <redis_port>, <redis_password>, and <redis_database> with your actual service name and Redis configuration details.

### Usage
Once configured, the plugin will automatically cache the responses of your service's API requests. The cache TTL and Redis settings can be adjusted as per your requirements.

