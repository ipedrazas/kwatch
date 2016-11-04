package main

type RdsEvent struct {
	Type   string `json:"type"`
	Object RdsDB  `json:"object"`
}

type RdsDB struct {
	ApiVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   map[string]string `json:"metadata"`
	Spec       RdsSpec           `json:"spec"`
}

type RdsSpec struct {
	Name          string `json:"name"`
	Engine        string `json:"engine"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	InstanceClass string `json:"instanceClass"`
	Storage       string `json:"storage"`
}
