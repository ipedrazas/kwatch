package main

// RdsEvent should be ApiEvent
type RdsEvent struct {
	Type   string
	Object RdsDB
}

// RdsDB should be Event
type RdsDB struct {
	APIVersion string
	Kind       string
	Metadata   map[string]string
	Spec       RdsSpec
}

// RdsSpec Should be PostgresSpec
type RdsSpec struct {
	Name          string
	Engine        string
	User          string
	Password      string
	InstanceClass string
	Storage       string
	Provider      string
	Service       string
	Status        string
}

// Webhook to send a response
type Webhook struct {
	Response      string
	URL           string
	TLSSkipVerify bool
	Debug         bool
}

// AuthConfig to configure how to connect to the APIServer
type AuthConfig struct {
	Token      string
	APIServer  string
	AuthBasic  AuthBasic
	Certs      AuthCerts
	SkipVerify bool
	Headers    map[string]string
	Watch      string
}

// AuthBasic object for basic auth
type AuthBasic struct {
	User     string
	Password string
	encoded  string
}

// AuthCerts to hold certificates details
type AuthCerts struct {
	ca   string
	cert string
	key  string
}

// Job models helm job entry
type Job struct {
	Image    string
	ImageTag string
	ID       string
}

// Helm models the wrapper to send data to Helm templates
type Helm struct {
	Values Values
}

// Values models the values.yaml entries
type Values struct {
	Job  Job
	Rds  RdsDB
	Test string
}
