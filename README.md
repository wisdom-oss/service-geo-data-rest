# WISdoM - Geo Data REST Adapter
Maintainer: [Jan Eike Suchard](mailto:jan.eike.suchard@uni-oldenburg.de)
<hr/>

## Information
This service is currently under heavy development and will only be available as development version until further notice

## Running the service
For running this service you need the following things:
- Docker
- (optional) Docker Compose
- A PostGIS-enabled PostgreSQL database

Furthermore, you need to set the following environment variables:
- AMQP_DSN
- SERVICE_REGISTRY_HOST
- DATABASE_DSN

#### Layer Configuration
To use this service you currently need to create a `layers.yaml` file which contains the 
configuration of the layers which are available for serving. The configuration file may look 
like [this](./layers.example.yaml):
<details>
    <summary>Click to see example file</summary>

```yaml
layers:
  - name: example-layer
    db_schema: public
    resolutions:
      - name: municipalities
        table: example_municipalities
      - name: districs
        table: example_districts
```
</details>

When writing your own configuration files you are able to use the following [JSON Schema](./layers.schema.json)
to validate your YAML file. (Yes, this is possible):

<details>
    <summary>Click to see the JSON Schema</summary>

```json
{
  "$schema": "http://json-schema.org/draft-07/schema",
  "title": "Layer Configuration Schema",
  "description": "The schema used to configure a Map Layer",
  "type": "array",
  "items": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string",
        "title": "Layer Name",
        "description": "The name of the layer is used as path parameter"
      },
      "db_schema": {
        "type": "string",
        "title": "Database Schema",
        "description": "The schema in which the tables for the resolutions are stored"
      },
      "resolutions": {
        "type": "array",
        "items": {
          "type": "object",
          "title": "Layer Resolution",
          "required": [
            "name", "table"
          ],
          "properties": {
            "name": {
              "type": "string",
              "title": "Resolution name",
              "description": "The name of the resolution which is used as path parameter"
            },
            "table": {
              "type": "string",
              "title": "Table name",
              "description": "The name of the database table which stores the geospatial data"
            },
            "default": {
              "type": "boolean",
              "default": false,
              "title": "Default Resolution",
              "description": "Indicator which is used if no resolution is given as path parameter"
            }
          }
        }
      }
    }
  }
}
```
</details>

