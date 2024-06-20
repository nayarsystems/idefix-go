package messages

const (
	// TopicTransportLogin is the login function
	TopicTransportLogin = TopicTransportIdefixPrefix + CmdLogin

	// TopicTransportDomainCreate is the cereate domain function
	TopicTransportDomainCreate = TopicTransportIdefixPrefix + CmdDomainCreate

	// TopicTransportDomainUpdate is the update domain function
	TopicTransportDomainUpdate = TopicTransportIdefixPrefix + CmdDomainUpdate

	// TopicTransportDomainUpdateAccessRules is the update domain function
	TopicTransportDomainUpdateAccessRules = TopicTransportIdefixPrefix + CmdDomainUpdateAccessRules

	// TopicTransportDomainDelete is the delete domain function
	TopicTransportDomainDelete = TopicTransportIdefixPrefix + CmdDomainDelete

	// TopicTransportDomainGet is the get domain function
	TopicTransportDomainGet = TopicTransportIdefixPrefix + CmdDomainGet

	// TopicTransportDomainAssign is the assign domain function
	TopicTransportDomainAssign = TopicTransportIdefixPrefix + CmdDomainAssign

	// TopicTransportDomainListAddresses list every address assigned to domains under the specified domain
	TopicTransportDomainListAddresses = TopicTransportIdefixPrefix + CmdDomainListAddresses

	// TopicTransportDomainListGroups: used to list every group under the specified domain
	TopicTransportDomainListGroups = TopicTransportIdefixPrefix + CmdDomainListGroups

	// TopicTransportDomainCountAddresses count every address assigned to domains under the specified domain
	TopicTransportDomainCountAddresses = TopicTransportIdefixPrefix + CmdDomainCountAddresses

	// TopicTransportDomainTree list every nested domain under the specified domain
	TopicTransportDomainTree = TopicTransportIdefixPrefix + CmdDomainTree

	// TopicTransportAddToGroup: add an address to a group
	TopicTransportAddToGroup = TopicTransportIdefixPrefix + CmdGroupAddAddress

	// TopicTransportRemoveFromGroup remove an address from a group
	TopicTransportRemoveFromGroup = TopicTransportIdefixPrefix + CmdGroupRemoveAddress

	// TopicTransportGetAddressGroups: add an address to a group
	TopicTransportGetAddressGroups = TopicTransportIdefixPrefix + CmdAddressGetGroups

	// TopicTransportGetAddressGroups: list all the addresses from a group
	TopicTransportGetGroupAddresses = TopicTransportIdefixPrefix + CmdGroupGetAddresses

	// TopicTransportEnvGet is the environment get function
	TopicTransportEnvGet = TopicTransportIdefixPrefix + CmdEnvGet

	// TopicTransportAddressStatesGet is used to get last state map of a client
	TopicTransportAddressStatesGet = TopicTransportIdefixPrefix + CmdAddressStatesGet

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

	// TopicTransportAddressConfigGet is used to get the custom data for the given client
	TopicTransportAddressConfigGet = TopicTransportIdefixPrefix + CmdAddressConfigGet

	// TopicTransportAddressConfigUpdate is used to update the custom data for the given client
	TopicTransportAddressConfigUpdate = TopicTransportIdefixPrefix + CmdAddressConfigUpdate

	// TopicTransportEventsCreate is used to create a new event on the server linked to the source address' domain
	TopicTransportEventsCreate = TopicTransportIdefixPrefix + CmdEventsCreate

	// TopicTransportEventsGet is used to get all events from a domain, since a timestamp
	TopicTransportEventsGet = TopicTransportIdefixPrefix + CmdEventsGet

	// TopicTransportEventsCreate is used to create a new event on the server linked to the source address' domain
	TopicTransportSchemasCreate = TopicTransportIdefixPrefix + CmdSchemasCreate

	// TopicTransportSchemasGet is used to get one schema
	TopicTransportSchemasGet = TopicTransportIdefixPrefix + CmdSchemasGet

	// TopicTransportSessionDelete is used to delete a session (logout)
	TopicTransportSessionDelete = TopicTransportIdefixPrefix + CmdSessionDelete

	// Mqtt prefix for idefix project
	MqttIdefixPrefix = "ifx"

	// Mqtt prefix for public messages
	MqttPublicPrefix = "p"

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

	// CmdDomainUpdateAccessRules is the update domain's access rules function
	CmdDomainUpdateAccessRules = "domain.rules.update"

	// CmdDomainDelete is the delete domain function
	CmdDomainDelete = "domain.delete"

	// CmdDomainGet is the get domain function
	CmdDomainGet = "domain.get"

	// CmdDomainAssign is the assign domain function
	CmdDomainAssign = "domain.assign"

	// CmdDomainListAddresses list every address assigned to domains under the specified domain
	CmdDomainListAddresses = "domain.list.addresses"

	// CmdDomainListGroups list every group under the specified domain
	CmdDomainListGroups = "domain.list.groups"

	// CmdDomainCountAddresses count every address assigned to domains under the specified domain
	CmdDomainCountAddresses = "domain.count.addresses"

	// CmdDomainTree list every nested domain under the specified domain
	CmdDomainTree = "domain.tree"

	// CmdGroupAddAddress is used to add an address to a group
	CmdGroupAddAddress = "group.add"

	// CmdGroupRemoveAddress is used to remove an address from a group
	CmdGroupRemoveAddress = "group.remove"

	// CmdGroupGetAddresses is used to get all the addresseses from a group
	CmdGroupGetAddresses = "group.get"

	// CmdEnvGet is the environment get function
	CmdEnvGet = "env.get"

	// CmdAddressStatesGet is used to get the last state map of an address
	CmdAddressStatesGet = "address.states.get"

	// CmdAddressGetGroups is used to get the groups of an address
	CmdAddressGetGroups = "address.groups.get"

	// CmdLocalRulesGet is used to get the local rules for the given client
	CmdAddressRulesGet = "address.rules.get"

	// CmdLocalRulesUpdate is used to update the local rules for the given client
	CmdAddressRulesUpdate = "address.rules.update"

	// CmdLocalRulesUpdate is used to update the local rules for the given client
	CmdAddressTokenReset = "address.token.reset"

	// CmdAddressCustomDataGet is used to get the custom data for the given client
	CmdAddressConfigGet = "address.config.get"

	// CmdAddressCustomDataUpdate is used to update the custom data for the given client
	CmdAddressConfigUpdate = "address.config.update"

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

	CmdSessionDelete = "session.delete"
)
