/*
 * Kafka Service Fleet Manager
 *
 * Kafka Service Fleet Manager APIs that are used by internal services e.g kas-fleetshard operators.
 *
 * API version: 1.1.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package private

// ManagedKafkaAllOfMetadata struct for ManagedKafkaAllOfMetadata
type ManagedKafkaAllOfMetadata struct {
	Name        string                               `json:"name,omitempty"`
	Namespace   string                               `json:"namespace,omitempty"`
	Annotations ManagedKafkaAllOfMetadataAnnotations `json:"annotations,omitempty"`
}
