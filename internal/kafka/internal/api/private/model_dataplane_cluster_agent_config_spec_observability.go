/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager APIs that are used by internal services e.g kas-fleetshard operators.
 *
 * API version: 1.3.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package private

// DataplaneClusterAgentConfigSpecObservability Observability configurations
type DataplaneClusterAgentConfigSpecObservability struct {
	// Deprecated
	DeprecatedAccessToken *string `json:"accessToken,omitempty"`
	AccessToken           *string `json:"access_token,omitempty"`
	Channel               string  `json:"channel,omitempty"`
	Repository            string  `json:"repository,omitempty"`
	Tag                   string  `json:"tag,omitempty"`
}
