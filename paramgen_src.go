// Code generated by paramgen. DO NOT EDIT.
// Source: github.com/ConduitIO/conduit-commons/tree/main/paramgen

package mqtt

import (
	"github.com/conduitio/conduit-commons/config"
)

const (
	SourceConfigBroker   = "broker"
	SourceConfigClientId = "clientId"
	SourceConfigPassword = "password"
	SourceConfigPort     = "port"
	SourceConfigQos      = "qos"
	SourceConfigTopic    = "topic"
	SourceConfigUsername = "username"
)

func (SourceConfig) Parameters() map[string]config.Parameter {
	return map[string]config.Parameter{
		SourceConfigBroker: {
			Default:     "",
			Description: "",
			Type:        config.ParameterTypeString,
			Validations: []config.Validation{
				config.ValidationRequired{},
			},
		},
		SourceConfigClientId: {
			Default:     "mqtt_conduit_client",
			Description: "",
			Type:        config.ParameterTypeString,
			Validations: []config.Validation{},
		},
		SourceConfigPassword: {
			Default:     "",
			Description: "",
			Type:        config.ParameterTypeString,
			Validations: []config.Validation{
				config.ValidationRequired{},
			},
		},
		SourceConfigPort: {
			Default:     "1883",
			Description: "",
			Type:        config.ParameterTypeInt,
			Validations: []config.Validation{},
		},
		SourceConfigQos: {
			Default:     "0",
			Description: "",
			Type:        config.ParameterTypeInt,
			Validations: []config.Validation{},
		},
		SourceConfigTopic: {
			Default:     "#",
			Description: "",
			Type:        config.ParameterTypeString,
			Validations: []config.Validation{},
		},
		SourceConfigUsername: {
			Default:     "",
			Description: "",
			Type:        config.ParameterTypeString,
			Validations: []config.Validation{
				config.ValidationRequired{},
			},
		},
	}
}