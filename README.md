# Conduit Connector for mqtt

[Conduit](https://conduit.io) connector for mqtt.

## How to build?

Run `make build` to build the connector.

## Testing

Run `make test` to run all the unit tests. Run `make test-integration` to run the integration tests.

The Docker compose file at `test/docker-compose.yml` can be used to run the required resource locally.

## Source

A source connector pulls data from an external resource and pushes it to downstream resources via Conduit.

### Configuration

| name                  | description                           | required | default value |
|-----------------------|---------------------------------------|----------|---------------|
| `broker`              | mqtt broker to connect to             | true     |               |
| `username`            | username to use to connect            | true     |               |
| `password`            | password to use to connect            | true     |               |
| `port`                | port on the broker to connect to      | false    | 1883          |
| `qos`                 | the qos on the _receiving_ side       | false    | 0             |
| `clientId`            | client id for the mqtt client         | false    | mqtt_conduit_client             |
| `topic`               | topic to receive messages on. Can use [mqtt wildcard syntax](https://www.hivemq.com/blog/mqtt-essentials-part-5-mqtt-topics-best-practices/)           | false    | `#`           |


## Known Issues & Limitations
- Accepts only one topic literal. However, all wildcard syntax supported by mqtt is supported in the connector so by using wildcards you can effectively receive from multiple topics 
- The mqtt spec dictates that the clientId must be unique per broker. So if you have multiple connectors or multiple pipelines you'll need to uniquely set the `clientId`


## Planned work
- [ ] Destination connector
