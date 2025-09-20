package messagequeue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/rabbitmq/amqp091-go"
)

const (
	defaultRetryInterval = 5 * time.Second
	defaultRetryMax      = 3
)

// Config holds the RabbitMQ configuration parameters.
type Config struct {
	Host             string `yaml:"host" json:"host"`
	Port             string `yaml:"port" json:"port"`
	User             string `yaml:"user" json:"user"`
	Password         string `yaml:"password" json:"password"`
	Vhost            string `yaml:"vhost" json:"vhost"`
	EnabledTLS       bool   `yaml:"enabled_tls" json:"enabled_tls"`
	ConnectionString string `yaml:"connection_string" json:"connection_string"`
}

func (c *Config) Validate() error {

	if c == nil {
		return fmt.Errorf("error validate message queue configurarion")
	}

	if c.ConnectionString == "" && c.Host == "" {
		return fmt.Errorf("connection strIng or host parameter must be configured")
	}

	if c.Host != "" {
		if c.Vhost == "" {
			return fmt.Errorf("vhost parameter must be configured")
		}

		if c.EnabledTLS {
			if c.Port == "" {
				c.Port = "5671"
			}
		} else {
			if c.Port == "" {
				c.Port = "5672"
			}
		}
		if c.User == "" {
			return fmt.Errorf("user parameter must be configured")
		}
		if c.Password == "" {
			return fmt.Errorf("password parameter must be configured")
		}
	}

	return nil
}

// QueueConfig holds the configuration for a single queue.
type QueueConfig struct {
	Name          string            `yaml:"name" json:"name"`
	Durable       bool              `yaml:"durable" json:"durable"`
	AutoDelete    bool              `yaml:"auto_delete" json:"auto_delete"`
	Exclusive     bool              `yaml:"exclusive" json:"exclusive"`
	NoWait        bool              `yaml:"no_wait" json:"no_wait"`
	Args          amqp091.Table     `yaml:"args" json:"args"`
	Bindings      []BindingConfig   `yaml:"bindings" json:"bindings"`
	DeadLetter    *DeadLetterConfig `yaml:"dead_letter" json:"dead_letter"`
	PrefetchCount int               `yaml:"prefetch_count" json:"prefetch_count"`
	Priority      int               `yaml:"priority" json:"priority"`
}

// ExchangeConfig holds the configuration for a single exchange.
type ExchangeConfig struct {
	Name        string          `yaml:"name" json:"name"`
	Kind        string          `yaml:"kind" json:"kind"`
	Durable     bool            `yaml:"durable" json:"durable"`
	AutoDeleted bool            `yaml:"auto_deleted" json:"auto_deleted"`
	Internal    bool            `yaml:"internal" json:"internal"`
	NoWait      bool            `yaml:"no_wait" json:"no_wait"`
	Args        amqp091.Table   `yaml:"args" json:"args"`
	Bindings    []BindingConfig `yaml:"bindings" json:"bindings"`
	Priority    int             `yaml:"priority" json:"priority"`
	Mandatory   bool            `yaml:"mandatory" json:"mandatory"`
	Immediate   bool            `yaml:"immediate" json:"immediate"`
}

// BindingConfig holds the configuration for a binding between an exchange and a queue.
type BindingConfig struct {
	QueueName    string        `yaml:"queue" json:"queue"`
	ExchangeName string        `yaml:"exchange" json:"exchange"`
	RoutingKey   string        `yaml:"routing_key" json:"routing_key"`
	NoWait       bool          `yaml:"no_wait" json:"no_wait"`
	Args         amqp091.Table `yaml:"args" json:"args"`
	Priority     int           `yaml:"priority" json:"priority"`
	Mandatory    bool          `yaml:"mandatory" json:"mandatory"`
	Immediate    bool          `yaml:"immediate" json:"immediate"`
}

// DeadLetterConfig holds the configuration for a dead-letter exchange and routing key.
type DeadLetterConfig struct {
	Exchange string `yaml:"exchange" json:"exchange"`
	Key      string `yaml:"routing_key" json:"routing_key"`
}

// FullConfig represents the complete RabbitMQ configuration, including queues, exchanges, and bindings.
type FullConfig struct {
	RabbitMQ  Config           `yaml:"config" json:"config"`
	Queues    []QueueConfig    `yaml:"queues" json:"queues"`
	Exchanges []ExchangeConfig `yaml:"exchanges" json:"exchanges"`
}

// Client is a RabbitMQ client.
type Client struct {
	config        Config
	conn          *amqp091.Connection
	channel       *amqp091.Channel
	queues        map[string]QueueConfig
	exchanges     map[string]ExchangeConfig
	fullConfig    FullConfig
	retryInterval time.Duration
	retryMax      int
	mu            sync.Mutex
	tracer        trace.Tracer
	propagator    propagation.TextMapPropagator
	consumers     map[string]context.CancelFunc
	consumersWg   sync.WaitGroup
	shutdown      chan struct{}
}

type MessageQueue interface {
	Initialize(fullConfig *FullConfig) error
	Connect(ctx context.Context) error
	Disconnect() error
	GracefulShutdown(ctx context.Context) error
	Publish(ctx context.Context, exchangeName string, routingKey string, publishing amqp091.Publishing) error
	Consume(ctx context.Context, queueName string, autoAck bool, consumerTag string, handler func(amqp091.Delivery) error) error
	GetQueueConfig(queueName string) (QueueConfig, bool)
	GetExchangeConfig(exchangeName string) (ExchangeConfig, bool)
	CancelConsumer(consumerTag string) error
}

// NewClient creates a new RabbitMQ client.
func NewClient(tracer trace.Tracer, propagator propagation.TextMapPropagator) MessageQueue {
	return &Client{
		queues:        make(map[string]QueueConfig),
		exchanges:     make(map[string]ExchangeConfig),
		consumers:     make(map[string]context.CancelFunc),
		shutdown:      make(chan struct{}),
		retryInterval: defaultRetryInterval,
		retryMax:      defaultRetryMax,
		tracer:        tracer,
		propagator:    propagator,
	}
}

// Initialize initializes the RabbitMQ client with configurations from a YAML file.
func (c *Client) Initialize(fullConfig *FullConfig) error {

	if fullConfig == nil {
		return fmt.Errorf("fullConfig is nil")
	}

	c.fullConfig = *fullConfig

	// Set the RabbitMQ configuration
	c.config = c.fullConfig.RabbitMQ

	// Initialize queues and exchanges
	for _, queue := range c.fullConfig.Queues {
		c.queues[queue.Name] = queue
	}

	for _, exchange := range c.fullConfig.Exchanges {
		c.exchanges[exchange.Name] = exchange
	}

	return nil
}

// Implementação do CancelConsumer (exemplo usando amqp091)
func (c *Client) CancelConsumer(consumerTag string) error {
	if c.channel == nil {
		return nil
	}

	// Cancela o consumer específico
	return c.channel.Cancel(consumerTag, false)
}

// Connect connects to the RabbitMQ server.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	var dsn string
	if c.config.ConnectionString != "" {
		dsn = c.config.ConnectionString
	} else {
		protocol := "amqp"
		if c.config.EnabledTLS {
			protocol = "amqps"
		}
		dsn = fmt.Sprintf("%s://%s:%s@%s:%s/%s", protocol, c.config.User, c.config.Password, c.config.Host, c.config.Port, c.config.Vhost)
	}

	for i := 0; i <= c.retryMax; i++ {
		c.conn, err = amqp091.Dial(dsn)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}

		log.Printf("Attempt %d/%d failed to connect to RabbitMQ: %v", i+1, c.retryMax+1, err)
		if i < c.retryMax {
			select {
			case <-ctx.Done():
				return ctx.Err() // Return if context is cancelled
			case <-time.After(c.retryInterval):
				// Wait before retrying
			}
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", c.retryMax+1, err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare queues and exchanges based on configuration
	if err := c.declareQueuesAndExchanges(ctx); err != nil {
		return fmt.Errorf("failed to declare queues and exchanges: %w", err)
	}

	return nil
}

// Disconnect closes the connection to the RabbitMQ server.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	return nil
}

func (c *Client) GracefulShutdown(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Starting graceful shutdown of RabbitMQ...")

	// Sinaliza shutdown para todos os consumers
	close(c.shutdown)

	// Cancela todos os consumers ativos
	for consumerTag, cancel := range c.consumers {
		log.Printf("Cancelling consumer: %s", consumerTag)
		cancel()

		// Cancela o consumer no RabbitMQ
		if c.channel != nil && !c.channel.IsClosed() {
			if err := c.channel.Cancel(consumerTag, false); err != nil {
				log.Printf("Failed to cancel consumer %s: %v", consumerTag, err)
			}
		}
	}

	// Aguarda todos os consumers terminarem ou timeout
	done := make(chan struct{})
	go func() {
		c.consumersWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All consumers stopped gracefully")
	case <-ctx.Done():
		log.Println("Shutdown timeout reached, forcing close")
		return ctx.Err()
	}

	// Fecha conexões
	return c.Disconnect()
}

// declareQueuesAndExchanges declares queues, exchanges, and bindings based on the configuration.
func (c *Client) declareQueuesAndExchanges(ctx context.Context) error {
	for _, exchangeConfig := range c.fullConfig.Exchanges {
		if err := c.declareExchange(ctx, exchangeConfig); err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchangeConfig.Name, err)
		}
	}

	for _, queueConfig := range c.fullConfig.Queues {
		if err := c.declareQueue(ctx, queueConfig); err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queueConfig.Name, err)
		}
	}

	for _, queueConfig := range c.fullConfig.Queues {
		for _, bindingConfig := range queueConfig.Bindings {
			// Assegura que o nome da queue está correto no binding
			binding := BindingConfig{
				QueueName:    queueConfig.Name,
				ExchangeName: bindingConfig.ExchangeName,
				RoutingKey:   bindingConfig.RoutingKey,
				NoWait:       bindingConfig.NoWait,
				Args:         bindingConfig.Args,
			}

			if err := c.bindQueue(ctx, binding); err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s with key %s: %w",
					binding.QueueName, binding.ExchangeName, binding.RoutingKey, err)
			}
		}
	}

	for _, exchangeConfig := range c.fullConfig.Exchanges {
		for _, bindingConfig := range exchangeConfig.Bindings {
			if err := c.bindQueue(ctx, bindingConfig); err != nil {
				return fmt.Errorf("failed to bind queue %s to exchange %s: %w", bindingConfig.QueueName, bindingConfig.ExchangeName, err)
			}
		}
	}

	return nil
}

// declareExchange declares an exchange.
func (c *Client) declareExchange(ctx context.Context, config ExchangeConfig) error {
	_, span := c.tracer.Start(ctx, "rabbitmq.declareExchange")
	defer span.End()

	err := c.channel.ExchangeDeclare(
		config.Name,        // name
		config.Kind,        // kind
		config.Durable,     // durable
		config.AutoDeleted, // autoDelete
		config.Internal,    // internal
		config.NoWait,      // noWait
		config.Args,        // args
	)
	if err != nil {
		return fmt.Errorf("exchange Declare: %w", err)
	}

	return nil
}

// declareQueue declares a queue.
func (c *Client) declareQueue(ctx context.Context, config QueueConfig) error {
	_, span := c.tracer.Start(ctx, "rabbitmq.declareQueue")
	defer span.End()

	_, err := c.channel.QueueDeclare(
		config.Name,       // name
		config.Durable,    // durable
		config.AutoDelete, // delete when unused
		config.Exclusive,  // exclusive
		config.NoWait,     // noWait
		config.Args,       // arguments
	)
	if err != nil {
		return fmt.Errorf("queue Declare: %w", err)
	}

	// Declare dead-letter exchange and queue if configured
	if config.DeadLetter != nil {
		deadLetterExchangeName := config.DeadLetter.Exchange
		deadLetterRoutingKey := config.DeadLetter.Key

		// Declare dead-letter exchange
		deadLetterExchangeConfig := ExchangeConfig{
			Name:        deadLetterExchangeName,
			Kind:        "direct", // You can adjust the kind as needed
			Durable:     true,
			AutoDeleted: false,
			Internal:    false,
			NoWait:      false,
			Args:        nil,
		}

		if err := c.declareExchange(ctx, deadLetterExchangeConfig); err != nil {
			return fmt.Errorf("failed to declare dead-letter exchange %s: %w", deadLetterExchangeName, err)
		}

		// Declare dead-letter queue
		deadLetterQueueName := config.Name + ".dead-letter"
		deadLetterQueueConfig := QueueConfig{
			Name:          deadLetterQueueName,
			Durable:       true,
			AutoDelete:    false,
			Exclusive:     false,
			NoWait:        false,
			Args:          nil,
			PrefetchCount: 1,
		}

		if err := c.declareQueue(ctx, deadLetterQueueConfig); err != nil {
			return fmt.Errorf("failed to declare dead-letter queue %s: %w", deadLetterQueueName, err)
		}

		// Bind dead-letter queue to dead-letter exchange
		deadLetterBindingConfig := BindingConfig{
			QueueName:    deadLetterQueueName,
			ExchangeName: deadLetterExchangeName,
			RoutingKey:   deadLetterRoutingKey,
			NoWait:       false,
			Args:         nil,
		}

		if err := c.bindQueue(ctx, deadLetterBindingConfig); err != nil {
			return fmt.Errorf("failed to bind dead-letter queue %s to exchange %s: %w", deadLetterQueueName, deadLetterExchangeName, err)
		}
	}

	return nil
}

// bindQueue binds a queue to an exchange.
func (c *Client) bindQueue(ctx context.Context, config BindingConfig) error {
	_, span := c.tracer.Start(ctx, "rabbitmq.bindQueue")
	defer span.End()

	err := c.channel.QueueBind(
		config.QueueName,    // name
		config.RoutingKey,   // key
		config.ExchangeName, // exchange
		config.NoWait,       // noWait
		config.Args,         // args
	)
	if err != nil {
		return fmt.Errorf("queue Bind: %w", err)
	}

	return nil
}

// Publish publishes a message to an exchange.
func (c *Client) Publish(ctx context.Context, exchangeName string, routingKey string, publishing amqp091.Publishing) error {
	ctx, span := c.tracer.Start(ctx, "rabbitmq.publish")
	defer span.End()
	if c == nil {
		return fmt.Errorf("client is nil")
	}

	if c.channel == nil {
		return fmt.Errorf("channel is nil")
	}
	// Verifica se o channel está disponível
	if c.channel == nil {
		err := fmt.Errorf("rabbitmq channel is nil")
		span.RecordError(err)
		return fmt.Errorf("exchange Publish: %w", err)
	}

	// Verifica se o channel não está fechado
	if c.channel.IsClosed() {
		err := fmt.Errorf("rabbitmq channel is closed")
		span.RecordError(err)
		return fmt.Errorf("exchange Publish: %w", err)
	}

	// Validações básicas
	if exchangeName == "" {
		err := fmt.Errorf("exchange name is empty")
		span.RecordError(err)
		return fmt.Errorf("exchange Publish: %w", err)
	}

	if routingKey == "" {
		err := fmt.Errorf("routing key is empty")
		span.RecordError(err)
		return fmt.Errorf("exchange Publish: %w", err)
	}

	if len(publishing.Body) == 0 {
		err := fmt.Errorf("message body is empty")
		span.RecordError(err)
		return fmt.Errorf("exchange Publish: %w", err)
	}
	err := c.channel.PublishWithContext(ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		publishing,   // publishing
	)
	if err != nil {
		return fmt.Errorf("exchange Publish: %w", err)
	}
	span.SetAttributes(attribute.String("status", "published_successfully"))
	return nil
}

// Consume consumes messages from a queue.
func (c *Client) Consume(ctx context.Context, queueName string, autoAck bool, consumerTag string, handler func(amqp091.Delivery) error) error {
	_, span := c.tracer.Start(ctx, "rabbitmq.consume")
	defer span.End()

	// Verifica se a conexão/channel estão válidos antes de consumir
	if c.conn == nil || c.conn.IsClosed() || c.channel == nil || c.channel.IsClosed() {
		if err := c.Connect(ctx); err != nil {
			return fmt.Errorf("failed to reconnect before consume: %w", err)
		}
	}

	msgs, err := c.channel.Consume(
		queueName,   // queue
		consumerTag, // consumer
		autoAck,     // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		return fmt.Errorf("queue Consume: %w", err)
	}

	// Cria context cancelável para este consumer
	consumerCtx, cancel := context.WithCancel(ctx)

	c.mu.Lock()
	c.consumers[consumerTag] = cancel
	c.consumersWg.Add(1)
	c.mu.Unlock()

	// Worker pool para processar mensagens sem criar goroutines ilimitadas
	const maxWorkers = 10 // Limite de workers concorrentes
	workCh := make(chan amqp091.Delivery, 100) // Buffer para mensagens pendentes

	// Inicia worker pool
	for i := 0; i < maxWorkers; i++ {
		c.consumersWg.Add(1)
		go func(workerID int) {
			defer c.consumersWg.Done()
			log.Printf("Worker %d started for consumer %s", workerID, consumerTag)

			for {
				select {
				case <-c.shutdown:
					log.Printf("Worker %d received shutdown signal", workerID)
					return
				case <-consumerCtx.Done():
					log.Printf("Worker %d context cancelled", workerID)
					return
				case delivery, ok := <-workCh:
					if !ok {
						log.Printf("Worker %d: work channel closed", workerID)
						return
					}

					// Processa mensagem com timeout
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("Worker %d: panic recovered: %v", workerID, r)
								if err := delivery.Nack(false, true); err != nil {
									log.Printf("Failed to Nack after panic: %v", err)
								}
							}
						}()

						if err := handler(delivery); err != nil {
							log.Printf("Worker %d: Handler error: %v", workerID, err)
							// Fallback NACK apenas se handler não controlou
							if err := delivery.Nack(false, true); err != nil {
								log.Printf("Worker %d: Failed to Nack: %v", workerID, err)
							}
						}
					}()
				}
			}
		}(i)
	}

	// Consumer principal - apenas distribui mensagens para workers
	go func() {
		defer func() {
			close(workCh) // Fecha canal para parar workers
			c.consumersWg.Done()
			c.mu.Lock()
			delete(c.consumers, consumerTag)
			c.mu.Unlock()
			log.Printf("Consumer %s stopped", consumerTag)
		}()

		for {
			select {
			case <-c.shutdown:
				log.Printf("Consumer %s received shutdown signal", consumerTag)
				return

			case <-consumerCtx.Done():
				log.Printf("Consumer %s context cancelled", consumerTag)
				return

			case d, ok := <-msgs:
				if !ok {
					log.Printf("Consumer %s: channel closed", consumerTag)
					return
				}

				// Envia para worker pool (com timeout para evitar bloqueio)
				select {
				case workCh <- d:
					// Mensagem enviada para processamento
				case <-time.After(5 * time.Second):
					log.Printf("Consumer %s: timeout sending to worker pool, nacking message", consumerTag)
					if err := d.Nack(false, true); err != nil {
						log.Printf("Failed to Nack timeout message: %v", err)
					}
				}
			}
		}
	}()

	log.Printf("Consumer %s started for queue %s", consumerTag, queueName)
	return nil
}

// GetQueueConfig returns the configuration for a queue.
func (c *Client) GetQueueConfig(queueName string) (QueueConfig, bool) {
	config, ok := c.queues[queueName]
	return config, ok
}

// GetExchangeConfig returns the configuration for an exchange.
func (c *Client) GetExchangeConfig(exchangeName string) (ExchangeConfig, bool) {
	config, ok := c.exchanges[exchangeName]
	return config, ok
}
