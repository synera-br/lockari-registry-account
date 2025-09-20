package messaging

type Queue string

const (
	QueueVaultCreateSecret     Queue = "q.vault.create_secret"
	QueueVaultReadSecret       Queue = "q.vault.read_secret"
	QueueVaultDeleteSecret     Queue = "q.vault.delete_secret"
	QueueVaultDeleteDLQSecret  Queue = "q.vault.delete_secret.dlq"
	QueueAuditEvents           Queue = "q.audit.events"
	QueueTenantCreateSecret    Queue = "q.tenant.create"
	QueueTenantReadSecret      Queue = "q.tenant.read"
	QueueTenantDeleteSecret    Queue = "q.tenant.delete"
	QueueTenantDeleteDLQSecret Queue = "q.tenant.delete.dlq"
	QueueUserCreateSecret      Queue = "q.user.create"
	QueueUserReadSecret        Queue = "q.user.read"
	QueueUserDeleteSecret      Queue = "q.user.delete"
	QueueUserDeleteDLQSecret   Queue = "q.user.delete.dlq"
	QueueAccountCreate         Queue = "q.account.create"
)

type RoutingKey string

const (
	RoutingKeyVaultCreateSecret     RoutingKey = "vault.create_secret"
	RoutingKeyVaultReadSecret       RoutingKey = "vault.read_secret"
	RoutingKeyVaultDeleteSecret     RoutingKey = "vault.delete_secret"
	RoutingKeyVaultDeleteDLQSecret  RoutingKey = "vault.delete_secret.dlq"
	RoutingKeyAuditEvents           RoutingKey = "audit.events"
	RoutingKeyTenantCreateSecret    RoutingKey = "tenant.create"
	RoutingKeyTenantReadSecret      RoutingKey = "tenant.read"
	RoutingKeyTenantDeleteSecret    RoutingKey = "tenant.delete"
	RoutingKeyTenantDeleteDLQSecret RoutingKey = "tenant.delete.dlq"
	RoutingKeyUserCreateSecret      RoutingKey = "user.create"
	RoutingKeyUserReadSecret        RoutingKey = "user.read"
	RoutingKeyUserDeleteSecret      RoutingKey = "user.delete"
	RoutingKeyUserDeleteDLQSecret   RoutingKey = "user.delete.dlq"
	RoutingKeyAccountCreate         RoutingKey = "account.create"
)

func (q Queue) String() string {
	return string(q)
}

type Exchange string

const (
	ExchangeUser            Exchange = "user_direct_exchange"
	ExchangeTenant          Exchange = "tenant_direct_exchange"
	ExchangeRegistryAccount Exchange = "registration_direct_exchange"
)

func (e Exchange) String() string {
	return string(e)
}

func (r RoutingKey) String() string {
	return string(r)
}
