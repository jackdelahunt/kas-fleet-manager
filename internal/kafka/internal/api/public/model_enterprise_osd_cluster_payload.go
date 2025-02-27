/*
 * Kafka Management API
 *
 * Kafka Management API is a REST API to manage Kafka instances
 *
 * API version: 1.15.0
 * Contact: rhosak-support@redhat.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package public

// EnterpriseOsdClusterPayload Schema for the request body sent to /clusters POST
type EnterpriseOsdClusterPayload struct {
	// Sets whether Kafkas created on this data plane cluster have to be accessed via private network
	AccessKafkasViaPrivateNetwork bool `json:"access_kafkas_via_private_network"`
	// The data plane cluster ID. This is the ID of the cluster obtained from OpenShift Cluster Manager (OCM) API
	ClusterId string `json:"cluster_id"`
	// dns name of the cluster. Can be obtained from the response JSON of the /api/clusters_mgmt/v1/clusters/<cluster_id>/ingresses (dns_name)
	ClusterIngressDnsName string `json:"cluster_ingress_dns_name"`
	// The node count given to the created kafka machine pool.  The machine pool must be created via /api/clusters_mgmt/v1/clusters/<cluster_id>/machine_pools prior to passing this value. The created machine pool must have a `bf2.org/kafkaInstanceProfileType=standard` label and a `bf2.org/kafkaInstanceProfileType=standard:NoExecute` taint. The name of the machine pool must be `kafka-standard`  The node count value has to be a multiple of 3 with a minimum of 3 nodes.
	KafkaMachinePoolNodeCount int32 `json:"kafka_machine_pool_node_count"`
}
