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

	// TopicTransportDomainEnvironmentGet is used to get the environment variables for a domain
	TopicTransportDomainEnvironmentGet = TopicTransportIdefixPrefix + CmdDomainEnvironmentGet

	// TopicTransportDomainEnvironmentSet is used to set the environment variables for a domain
	TopicTransportDomainEnvironmentSet = TopicTransportIdefixPrefix + CmdDomainEnvironmentSet

	// TopicTransportDomainEnvironmentUnset is used to unset the environment variables for a domain
	TopicTransportDomainEnvironmentUnset = TopicTransportIdefixPrefix + CmdDomainEnvironmentUnset

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

	// TopicTransportEnvironmentGet is the environment get function
	TopicTransportEnvironmentGet = TopicTransportIdefixPrefix + CmdEnvironmentGet

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

	// TopicTransportAddressEnvironmentGet is used to update the custom data for the given client
	TopicTransportAddressEnvironmentGet = TopicTransportIdefixPrefix + CmdAddressEnvironmentGet

	// TopicTransportAddressEnvironmentGet is used to update the custom data for the given client
	TopicTransportAddressEnvironmentSet = TopicTransportIdefixPrefix + CmdAddressEnvironmentSet

	// TopicTransportAddressEnvironmentGet is used to update the custom data for the given client
	TopicTransportAddressEnvironmentUnset = TopicTransportIdefixPrefix + CmdAddressEnvironmentUnset

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

	// TopicTransportAddressAliasGet is used to get the aliases assigned to an address
	TopicTransportAddressAliasGet = TopicTransportIdefixPrefix + CmdAddressAliasGet

	// TopicTransportAddressAliasAdd is used to add an alias to an address
	TopicTransportAddressAliasAdd = TopicTransportIdefixPrefix + CmdAddressAliasAdd

	// TopicTransportAddressAliasRemove is used to remove an alias from an address
	TopicTransportAddressAliasRemove = TopicTransportIdefixPrefix + CmdAddressAliasRemove

	// MqttIdefixPrefix represents the MQTT prefix for the Idefix project
	MqttIdefixPrefix = "ifx"

	// MqttPublicPrefix represents the MQTT prefix for public messages
	MqttPublicPrefix = "p"

	// IdefixCmdPrefix is the base prefix for all cloud system commands related to Idefix, such as login
	IdefixCmdPrefix = "idefix"

	// TopicTransportIdefixPrefix is the MQTT topic prefix for all cloud system commands under the Idefix namespace
	TopicTransportIdefixPrefix = IdefixCmdPrefix + "."

	// CmdLogin is the command for user login
	CmdLogin = "login"

	// CmdDomainCreate is the command to create a new domain
	CmdDomainCreate = "domain.create"

	// CmdDomainUpdate is the command to update an existing domain
	CmdDomainUpdate = "domain.update"

	// CmdDomainUpdateAccessRules is the command to update the access rules of a domain
	CmdDomainUpdateAccessRules = "domain.rules.update"

	// CmdDomainDelete is the command to delete a domain
	CmdDomainDelete = "domain.delete"

	// CmdDomainGet is the command to retrieve details of a specific domain
	CmdDomainGet = "domain.get"

	// CmdDomainAssign is the command to assign an address to a domain
	CmdDomainAssign = "domain.assign"

	// CmdDomainListAddresses is the command to list all addresses associated with a specific domain
	CmdDomainListAddresses = "domain.list.addresses"

	// CmdDomainListGroups is the command to list all groups within a specific domain
	CmdDomainListGroups = "domain.list.groups"

	// CmdDomainCountAddresses is the command to count the addresses associated with a specific domain
	CmdDomainCountAddresses = "domain.count.addresses"

	// CmdDomainTree is the command to list all nested domains under a specific domain
	CmdDomainTree = "domain.tree"

	// CmdDomainEnvironmentGet is the command to retrieve the environment variables for a specific domain
	CmdDomainEnvironmentGet = "domain.environment.get"

	// CmdDomainEnvironmentSet is the command to set the environment variables for a specific domain
	CmdDomainEnvironmentSet = "domain.environment.set"

	// CmdDomainEnvironmentUnset is the command to unset the environment variables for a specific domain
	CmdDomainEnvironmentUnset = "domain.environment.unset"

	// CmdGroupAddAddress is the command to add an address to a group on a domain
	CmdGroupAddAddress = "group.add"

	// CmdGroupRemoveAddress is the command to remove an address from a group on a domain
	CmdGroupRemoveAddress = "group.remove"

	// CmdGroupGetAddresses is the command to retrieve all addresses from a group on a domain
	CmdGroupGetAddresses = "group.get"

	// CmdEnvGet is the command to retrieve environment details
	CmdEnvGet         = "env.get"
	CmdEnvironmentGet = "environment.get"

	// CmdAddressStatesGet is the command to get the last known state map of a specific address
	CmdAddressStatesGet = "address.states.get"

	// CmdAddressGetGroups is the command to retrieve the groups associated with a specific address
	CmdAddressGetGroups = "address.groups.get"

	// CmdAddressRulesGet is the command to retrieve the local rules associated with a specific address
	CmdAddressRulesGet = "address.rules.get"

	// CmdAddressRulesUpdate is the command to update the local rules for a specific address
	CmdAddressRulesUpdate = "address.rules.update"

	// CmdAddressTokenReset is the command to reset the token for a specific address
	CmdAddressTokenReset = "address.token.reset"

	// CmdAddressConfigGet is the command to retrieve the custom configuration data for a specific address
	CmdAddressConfigGet = "address.config.get"

	// CmdAddressConfigUpdate is the command to update the custom configuration data for a specific address
	CmdAddressConfigUpdate = "address.config.update"

	// CmdAddressEnvironmentGet is the command to retrieve the environment variables for a specific address
	CmdAddressEnvironmentGet = "address.environment.get"

	// CmdAddressEnvironmentSet is the command to set the environment variables for a specific address
	CmdAddressEnvironmentSet = "address.environment.set"

	// CmdAddressEnvironmentUnset is the command to unset the environment variables for a specific address
	CmdAddressEnvironmentUnset = "address.environment.unset"

	// CmdAddressDisable is the command to restrict access to a specific address
	CmdAddressDisable = "address.disable"

	// CmdAddressDomainGet is the command to retrieve the domain associated with a specific address
	CmdAddressDomainGet = "address.domain.get"

	// CmdAddressAliasGet is the command to retrieve all aliases associated with a specific address
	CmdAddressAliasGet = "address.alias.get"

	// CmdAddressAliasAdd is the command to add a new alias to a specific address
	CmdAddressAliasAdd = "address.alias.add"

	// CmdAddressAliasRemove is the command to remove an alias from a specific address
	CmdAddressAliasRemove = "address.alias.remove"

	// CmdEventsCreate is the command to create a new event linked to the source address's domain
	CmdEventsCreate = "events.create"

	// CmdEventsGet is the command to retrieve all events from a domain, since a timestamp
	CmdEventsGet = "events.get"

	// CmdSchemasCreate is the command to create a new schema
	CmdSchemasCreate = "schemas.create"

	// CmdSchemasGet is the command to retrieve schemas from a domain
	CmdSchemasGet = "schemas.get"

	// CmdSessionDelete is the command to delete a session
	CmdSessionDelete = "session.delete"
)
