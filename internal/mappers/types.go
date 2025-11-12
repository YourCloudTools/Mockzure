package mappers

// StoreInterface defines the interface for accessing store data
// This allows mappers to work with the Store without tight coupling
type StoreInterface interface {
	GetResourceGroups() []interface{}
	GetVMs() []interface{}
	GetUsers() []interface{}
	GetServiceAccounts() []interface{}
}

