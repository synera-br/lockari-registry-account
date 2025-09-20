```

**Explanation and Usage Guide:**

1.  **Configuration Structures:**
    *   `Config`: Basic RabbitMQ connection details.
    *   `QueueConfig`: Defines queue properties like durability, auto-deletion, and dead-letter settings.
    *   `ExchangeConfig`: Defines exchange properties (name, kind, durability, etc.).
    *   `BindingConfig`: Specifies how queues are bound to exchanges using routing keys.
    *   `DeadLetterConfig`: Configuration for dead-letter exchanges and routing keys.
    *   `FullConfig`: Aggregates all configurations (RabbitMQ connection, queues, exchanges).

2.  **Client Structure:**
    *   `Client`: Holds the connection, channel, configurations, retry settings, and OpenTelemetry components.
    *   `NewClient`: Constructor to create a new `Client` instance.

3.  **Initialization:**
    *   `Initialize(configFile string)`: Reads the RabbitMQ configuration from a YAML file.  It unmarshals the YAML into the `FullConfig` struct and populates the `queues` and `exchanges` maps.

4.  **Connection Management:**
    *   `Connect(ctx context.Context)`: Establishes a connection to the RabbitMQ server with retry logic.  It also declares queues and exchanges based on the loaded configuration.
    *   `Disconnect()`: Closes the channel and connection to RabbitMQ.

5.  **Declaration Functions:**
    *   `declareQueuesAndExchanges(ctx context.Context)`: Declares all queues, exchanges, and bindings defined in the configuration.
    *   `declareExchange(ctx context.Context, config ExchangeConfig)`: Declares a single exchange.
    *   `declareQueue(ctx context.Context, config QueueConfig)`: Declares a single queue, including dead-letter queues if configured.
    *   `bindQueue(ctx context.Context, config BindingConfig)`: Binds a queue to an exchange.

6.  **Publishing and Consuming:**
    *   `Publish(ctx context.Context, exchangeName string, routingKey string, publishing amqp091.Publishing)`: Publishes a message to the specified exchange with the given routing key.
    *   `Consume(ctx context.Context, queueName string, autoAck bool, consumerTag string, handler func(amqp091.Delivery))`: Consumes messages from a specified queue and processes them using the provided handler function.