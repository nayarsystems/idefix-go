package messages

const (
	// TopicTransportLogin is the login function
	TopicTransportLogin = TopicTransportIdefixPrefix + CmdLogin

	// TopicTransportDomainCreate is the cereate domain function
	TopicTransportDomainCreate = TopicTransportIdefixPrefix + CmdDomainCreate

	// TopicTransportDomainUpdate is the update domain function
	TopicTransportDomainUpdate = TopicTransportIdefixPrefix + CmdDomainUpdate

	// TopicTransportDomainDelete is the delete domain function
	TopicTransportDomainDelete = TopicTransportIdefixPrefix + CmdDomainDelete

	// TopicTransportDomainGet is the get domain function
	TopicTransportDomainGet = TopicTransportIdefixPrefix + CmdDomainGet

	// TopicTransportDomainAssign is the assign domain function
	TopicTransportDomainAssign = TopicTransportIdefixPrefix + CmdDomainAssign

	// TopicTransportDomainListAddresses list every address assigned to domains under the specified domain
	TopicTransportDomainListAddresses = TopicTransportIdefixPrefix + CmdDomainListAddresses

	// TopicTransportDomainCountAddresses count every address assigned to domains under the specified domain
	TopicTransportDomainCountAddresses = TopicTransportIdefixPrefix + CmdDomainCountAddresses

	// TopicTransportDomainTree list every nested domain under the specified domain
	TopicTransportDomainTree = TopicTransportIdefixPrefix + CmdDomainTree

	// TopicTransportEnvGet is the environment get function
	TopicTransportEnvGet = TopicTransportIdefixPrefix + CmdEnvGet

	// TopicTransportLocalRulesGet is used to get the local rules for the given client
	TopicTransportAddressRulesGet = TopicTransportIdefixPrefix + CmdAddressRulesGet

	// TopicTransportLocalRulesUpdate is used to update the local rules for the given client
	TopicTransportAddressRulesUpdate = TopicTransportIdefixPrefix + CmdAddressRulesUpdate

	// TopicTransportLocalRulesUpdate is used to update the local rules for the given client
	TopicTransportAddressTokenReset = TopicTransportIdefixPrefix + CmdAddressTokenReset

	// TopicTransportAddressDisable is used to prevent further activity from an address
	TopicTransportAddressDisable = TopicTransportIdefixPrefix + CmdAddressDisable

	// TopicTransportAddressDomainGet is used to get the domain assigned to address
	TopicTransportAddressDomainGet = TopicTransportIdefixPrefix + CmdAddressDomainGet

	// TopicTransportEventsCreate is used to create a new event on the server linked to the source address' domain
	TopicTransportEventsCreate = TopicTransportIdefixPrefix + CmdEventsCreate

	// TopicTransportEventsGet is used to get all events from a domain, since a timestamp
	TopicTransportEventsGet = TopicTransportIdefixPrefix + CmdEventsGet

	// TopicTransportEventsCreate is used to create a new event on the server linked to the source address' domain
	TopicTransportSchemasCreate = TopicTransportIdefixPrefix + CmdSchemasCreate

	// TopicTransportEventsGet is used to get all events from a domain, since a timestamp
	TopicTransportSchemasGet = TopicTransportIdefixPrefix + CmdSchemasGet

	// Mqtt prefix for idefix project
	MqttIdefixPrefix = "ifx"

	// IdefixPrefix is the prefix for all cloud system commands like login
	IdefixCmdPrefix = "idefix"

	// TopicTransportIdefixPrefix is the prefix for all cloud system commands like login
	TopicTransportIdefixPrefix = IdefixCmdPrefix + "."

	// CmdLogin is the login function
	CmdLogin = "login"

	// CmdDomainCreate is the cereate domain function
	CmdDomainCreate = "domain.create"

	// CmdDomainUpdate is the update domain function
	CmdDomainUpdate = "domain.update"

	// CmdDomainDelete is the delete domain function
	CmdDomainDelete = "domain.delete"

	// CmdDomainGet is the get domain function
	CmdDomainGet = "domain.get"

	// CmdDomainAssign is the assign domain function
	CmdDomainAssign = "domain.assign"

	// CmdDomainListAddresses list every address assigned to domains under the specified domain
	CmdDomainListAddresses = "domain.list.addresses"

	// CmdDomainCountAddresses count every address assigned to domains under the specified domain
	CmdDomainCountAddresses = "domain.count.addresses"

	// CmdDomainTree list every nested domain under the specified domain
	CmdDomainTree = "domain.tree"

	// CmdEnvGet is the environment get function
	CmdEnvGet = "env.get"

	// CmdLocalRulesGet is used to get the local rules for the given client
	CmdAddressRulesGet = "address.rules.get"

	// CmdLocalRulesUpdate is used to update the local rules for the given client
	CmdAddressRulesUpdate = "address.rules.update"

	// CmdLocalRulesUpdate is used to update the local rules for the given client
	CmdAddressTokenReset = "address.token.reset"

	// CmdAddressDisable is used to restrict the access to an address
	CmdAddressDisable = "address.disable"

	// CmdAddressDomainGet is used to get the domain assigned to an address
	CmdAddressDomainGet = "address.domain.get"

	// CmdEventsCreate is used to create a new event on the server linked to the source address' domain
	CmdEventsCreate = "events.create"

	// CmdEventsGet is used to get all events from a domain, since a timestamp
	CmdEventsGet = "events.get"

	// CmdEventsCreate is used to create a new event on the server linked to the source address' domain
	CmdSchemasCreate = "schemas.create"

	// CmdEventsGet is used to get all events from a domain, since a timestamp
	CmdSchemasGet = "schemas.get"
)
