package services

import (
	"context"
	"reflect"
	"testing"

	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/api"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/config"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/errors"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/ocm"
	clustersmgmtv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

func Test_DataPlaneCluster_UpdateDataPlaneClusterStatus(t *testing.T) {
	testClusterID := "test-cluster-id"
	tests := []struct {
		name                           string
		clusterID                      string
		clusterStatus                  *api.DataPlaneClusterStatus
		dataPlaneClusterServiceFactory func() *dataPlaneClusterService
		wantErr                        bool
	}{
		{
			name:          "An error is returned when a non-existent ClusterID is passed",
			clusterID:     testClusterID,
			clusterStatus: nil,
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				ocmClient := &ocm.ClientMock{
					GetClusterFunc: func(clusterID string) (*v1.Cluster, error) {
						return &v1.Cluster{}, nil
					},
				}
				clusterService := &ClusterServiceMock{
					FindClusterByIDFunc: func(clusterID string) (*api.Cluster, *errors.ServiceError) {
						return nil, nil
					},
				}
				return NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
			},
			wantErr: true,
		},
		{
			name:      "It succeeds when there are no issues",
			clusterID: testClusterID,
			clusterStatus: &api.DataPlaneClusterStatus{
				Conditions: []api.DataPlaneClusterStatusCondition{
					api.DataPlaneClusterStatusCondition{
						Type:   "Ready",
						Status: "True",
					},
				},
				NodeInfo: api.DataPlaneClusterStatusNodeInfo{
					Current: 6,
				},
			},
			wantErr: false,
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {

				ocmClient := &ocm.ClientMock{
					GetClusterFunc: func(clusterID string) (*v1.Cluster, error) {
						clusterBuilder := clustersmgmtv1.NewCluster()
						clusterNodeBuilder := clustersmgmtv1.NewClusterNodes()
						clusterNodeBuilder.Compute(6)
						clusterMetricsBuilder := clustersmgmtv1.NewClusterMetrics()
						clusterMetricsBuilder.Nodes(clusterNodeBuilder)
						clusterBuilder.Metrics(clusterMetricsBuilder)
						clusterBuilder.Nodes(clusterNodeBuilder)
						return clusterBuilder.Build()
					},
				}
				clusterService := &ClusterServiceMock{
					FindClusterByIDFunc: func(clusterID string) (*api.Cluster, *errors.ServiceError) {
						return &api.Cluster{
							Meta: api.Meta{
								ID: "id",
							},
							ClusterID: clusterID,
							Status:    api.ClusterReady,
						}, nil
					},
					UpdateStatusFunc: func(cluster api.Cluster, status api.ClusterStatus) error {
						return nil
					},
				}
				return NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataPlaneClusterService := tt.dataPlaneClusterServiceFactory()
			svcErr := dataPlaneClusterService.UpdateDataPlaneClusterStatus(context.Background(), tt.clusterID, tt.clusterStatus)
			gotErr := svcErr != nil
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("UpdateDataPlaneClusterStatus() error = %v, wantErr = %v", svcErr, tt.wantErr)
			}
		})
	}
}

func Test_DataPlaneCluster_updateDataPlaneClusterNodes(t *testing.T) {
	testClusterID := "test-cluster-id"

	type input struct {
		status                  *api.DataPlaneClusterStatus
		cluster                 *api.Cluster
		dataPlaneClusterService *dataPlaneClusterService
	}
	cases := []struct {
		name           string
		inputFactory   func() *input
		expectedResult int
		wantErr        bool
	}{
		{
			name: "when scale-up thresholds are crossed number of compute nodes is increased",
			inputFactory: func() *input {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 3
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
				}
				kafkaConfig := sampleValidConfig()
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, kafkaConfig)
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 6,
			wantErr:        false,
		},
		{
			name: "when a single scale-up threshold is crossed number of compute nodes is increased",
			inputFactory: func() *input {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 3
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.Remaining.Connections = 10000000000
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 6,
			wantErr:        false,
		},
		{
			name: "when scale-up threshold is crossed but scale-up nodes would be higher than restricted celing then no scaling is performed",
			inputFactory: func() *input {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 3
				testStatus.NodeInfo.Ceiling = 5 // We test restricted ceiling rounding here
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 3,
			wantErr:        false,
		},
		{
			name: "when all scale-down threshold is crossed number of compute nodes is decreased",
			inputFactory: func() *input {
				kafkaConfig := sampleValidConfig().Kafka
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 6
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				// We set remaining to a value much higher than resizeInfo.value which to
				// simulate a scale-down is needed, as scale-down thresholds are
				// calculated from resizeInfo.Delta value
				testStatus.ResizeInfo.Delta.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 10
				testStatus.ResizeInfo.Delta.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 10
				testStatus.Remaining.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 1000
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 1000
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
				}

				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 3,
			wantErr:        false,
		},
		{
			name: "when not all scale-down threshold are crossed number of compute nodes is not decreased",
			inputFactory: func() *input {
				kafkaConfig := sampleValidConfig().Kafka
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 6
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.ResizeInfo.Delta.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 10
				testStatus.ResizeInfo.Delta.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 10
				// We simulate connections scale-down threshold not being crossed
				// and partitions scale-down threshold being crossed
				testStatus.Remaining.Connections = testStatus.ResizeInfo.Delta.Connections - 1
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 1000
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 6,
			wantErr:        false,
		},
		{
			name: "when scale-down threshold is crossed but scaled-down nodes would be less than workloadMin then no scaling is performed",
			inputFactory: func() *input {
				kafkaConfig := sampleValidConfig().Kafka
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 6
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 6
				// We set remaining to a value much higher than resizeInfo.value which to
				// simulate a scale-down is needed, as scale-down thresholds are
				// calculated from resizeInfo.Delta value
				testStatus.ResizeInfo.Delta.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 10
				testStatus.ResizeInfo.Delta.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 10
				testStatus.Remaining.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 1000
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 1000
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 6,
			wantErr:        false,
		},
		{
			name: "when scale-down threshold is crossed but scaled-down nodes would be less than restricted floor then no scaling is performed",
			inputFactory: func() *input {
				kafkaConfig := sampleValidConfig().Kafka
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 6
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.NodeInfo.Floor = 5 // We test the rounding of restricted floor here
				// We set remaining to a value much higher than resizeInfo.value which to
				// simulate a scale-down is needed, as scale-down thresholds are
				// calculated from resizeInfo.Delta value
				testStatus.ResizeInfo.Delta.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 10
				testStatus.ResizeInfo.Delta.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 10
				testStatus.Remaining.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 1000
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 1000
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 6,
			wantErr:        false,
		},
		{
			name: "when no scale-up or scale-down thresholds are crossed no scaling is performed",
			inputFactory: func() *input {
				kafkaConfig := sampleValidConfig().Kafka
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				testStatus.NodeInfo.Current = 12
				testStatus.NodeInfo.Ceiling = 30
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.NodeInfo.Floor = 3

				// We set remaining higher than a single kafka instance capacity to not
				// trigger scale-up and we set it less than delta values to not force a
				// scale-down
				testStatus.Remaining.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 2
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 2
				testStatus.ResizeInfo.Delta.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections * 10
				testStatus.ResizeInfo.Delta.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions * 10

				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						return nil, nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}
			},
			expectedResult: 12,
			wantErr:        false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.inputFactory()
			if input == nil {
				t.Fatalf("invalid input")
			}

			dataPlaneClusterService := input.dataPlaneClusterService
			nodesAfterScaling, err := dataPlaneClusterService.updateDataPlaneClusterNodes(input.cluster, input.status)

			if !reflect.DeepEqual(nodesAfterScaling, tt.expectedResult) {
				t.Errorf("updateDataPlaneClusterNodes() got = %+v, expected %+v", nodesAfterScaling, tt.expectedResult)
			}
			if !reflect.DeepEqual(err != nil, tt.wantErr) {
				t.Errorf("updateDataPlaneClusterNodes() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_DataPlaneCluster_computeNodeScalingActionInProgress(t *testing.T) {
	testClusterID := "test-cluster-id"
	tests := []struct {
		name                           string
		clusterStatus                  *api.DataPlaneClusterStatus
		dataPlaneClusterServiceFactory func() *dataPlaneClusterService
		wantErr                        bool
		want                           bool
	}{
		{
			name:          "When desired compute nodes equals existing compute nodes no scaling action is in progress",
			clusterStatus: nil,
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				clusterBuilder := clustersmgmtv1.NewCluster()
				clusterBuilder.ID(testClusterID)
				clusterNodeBuilder := clustersmgmtv1.NewClusterNodes()
				clusterNodeBuilder.Compute(6)
				clusterMetricsBuilder := clustersmgmtv1.NewClusterMetrics()
				clusterMetricsBuilder.Nodes(clusterNodeBuilder)
				clusterBuilder.Metrics(clusterMetricsBuilder)
				clusterBuilder.Nodes(clusterNodeBuilder)
				cluster, err := clusterBuilder.Build()
				if err != nil {
					return nil
				}

				ocmClient := &ocm.ClientMock{
					GetClusterFunc: func(clusterID string) (*v1.Cluster, error) {
						return cluster, nil
					},
				}
				clusterService := &ClusterServiceMock{}
				return NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
			},
			want:    false,
			wantErr: false,
		},
		{
			name:          "When desired compute nodes does not equal existing compute nodes scaling action is in progress",
			clusterStatus: nil,
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				clusterBuilder := clustersmgmtv1.NewCluster()
				clusterBuilder.ID(testClusterID)
				clusterNodeBuilder := clustersmgmtv1.NewClusterNodes()
				clusterNodeBuilder.Compute(6)
				clusterBuilder.Nodes(clusterNodeBuilder)
				clusterNodeBuilderExisting := clustersmgmtv1.NewClusterNodes()
				clusterNodeBuilderExisting.Compute(8)
				clusterMetricsBuilder := clustersmgmtv1.NewClusterMetrics()
				clusterMetricsBuilder.Nodes(clusterNodeBuilderExisting)
				clusterBuilder.Metrics(clusterMetricsBuilder)
				cluster, err := clusterBuilder.Build()
				if err != nil {
					return nil
				}

				ocmClient := &ocm.ClientMock{
					GetClusterFunc: func(clusterID string) (*v1.Cluster, error) {
						return cluster, nil
					},
				}
				clusterService := &ClusterServiceMock{}
				return NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
			},
			want:    true,
			wantErr: false,
		},
		{
			name:          "When some node information is missing an error is returned",
			clusterStatus: nil,
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				clusterBuilder := clustersmgmtv1.NewCluster()
				clusterBuilder.ID(testClusterID)
				clusterNodeBuilder := clustersmgmtv1.NewClusterNodes()
				clusterNodeBuilder.Compute(6)
				clusterBuilder.Nodes(clusterNodeBuilder)
				cluster, err := clusterBuilder.Build()
				if err != nil {
					return nil
				}

				ocmClient := &ocm.ClientMock{
					GetClusterFunc: func(clusterID string) (*v1.Cluster, error) {
						return cluster, nil
					},
				}
				clusterService := &ClusterServiceMock{}
				return NewDataPlaneClusterService(clusterService, ocmClient, sampleValidConfig())
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.dataPlaneClusterServiceFactory()
			if f == nil {
				t.Fatalf("dataPlaneClusterService is nil")
			}

			testAPICluster := &api.Cluster{
				ClusterID: testClusterID,
			}
			res, err := f.computeNodeScalingActionInProgress(testAPICluster, nil)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("computeNodeScalingActionInProgress() got = %+v, expected %+v", res, tt.want)
			}
			if !reflect.DeepEqual(err != nil, tt.wantErr) {
				t.Errorf("computeNodeScalingActionInProgress() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_DataPlaneCluster_isFleetShardOperatorReady(t *testing.T) {
	tests := []struct {
		name                           string
		clusterStatus                  *api.DataPlaneClusterStatus
		dataPlaneClusterServiceFactory func() *dataPlaneClusterService
		wantErr                        bool
		want                           bool
	}{
		{
			name: "When KAS Fleet operator reports ready condition set to true the fleet shard operator is considered ready",
			clusterStatus: &api.DataPlaneClusterStatus{
				Conditions: []api.DataPlaneClusterStatusCondition{
					api.DataPlaneClusterStatusCondition{
						Type:   "Ready",
						Status: "True",
					},
				},
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "When KAS Fleet operator reports ready condition set to false the fleet shard operator is considered not ready",
			clusterStatus: &api.DataPlaneClusterStatus{
				Conditions: []api.DataPlaneClusterStatusCondition{
					api.DataPlaneClusterStatusCondition{
						Type:   "Ready",
						Status: "False",
					},
				},
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			wantErr: false,
			want:    false,
		},
		{
			name: "When KAS Fleet operator reports doesn't report a Ready condition the fleet shard operator is considered not ready",
			clusterStatus: &api.DataPlaneClusterStatus{
				Conditions: []api.DataPlaneClusterStatusCondition{},
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			wantErr: false,
			want:    false,
		},
		{
			name: "When KAS Fleet operator reports reports a Ready condition with an unknown value an error is returned",
			clusterStatus: &api.DataPlaneClusterStatus{
				Conditions: []api.DataPlaneClusterStatusCondition{
					api.DataPlaneClusterStatusCondition{
						Type:   "Ready",
						Status: "InventedValue",
					},
				},
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			wantErr: true,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.dataPlaneClusterServiceFactory()
			if f == nil {
				t.Fatalf("dataPlaneClusterService is nil")
			}

			res, err := f.isFleetShardOperatorReady(tt.clusterStatus)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("isFleetShardOperatorReady() got = %+v, expected %+v", res, tt.want)
			}
			if !reflect.DeepEqual(err != nil, tt.wantErr) {
				t.Errorf("isFleetShardOperatorReady() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func Test_DataPlaneCluster_clusterCanProcessStatusReports(t *testing.T) {
	tests := []struct {
		name                           string
		apiCluster                     *api.Cluster
		dataPlaneClusterServiceFactory func() *dataPlaneClusterService
		want                           bool
	}{
		{
			name: "When cluster is ready then status reports can be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterReady,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: true,
		},
		{
			name: "When cluster is full then status reports can be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterFull,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: true,
		},
		{
			name: "When cluster is waiting for KAS Fleet Shard operator then status reports can be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterWaitingForKasFleetShardOperator,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: true,
		},
		{
			name: "When cluster is in state scaling up then status reports cannot be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterComputeNodeScalingUp,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: true,
		},
		{
			name: "When cluster is in state provisioning then status reports cannot be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterProvisioning,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: false,
		},
		{
			name: "When cluster is in state failed then status reports cannot be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterFailed,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: false,
		},
		{
			name: "When cluster is in state accepted then status reports cannot be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterAccepted,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: false,
		},
		{
			name: "When cluster is in state provisioned then status reports cannot be processed",
			apiCluster: &api.Cluster{
				Status: api.ClusterProvisioned,
			},
			dataPlaneClusterServiceFactory: func() *dataPlaneClusterService {
				return NewDataPlaneClusterService(nil, nil, &config.ApplicationConfig{})
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.dataPlaneClusterServiceFactory()
			if f == nil {
				t.Fatalf("dataPlaneClusterService is nil")
			}

			res := f.clusterCanProcessStatusReports(tt.apiCluster)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("clusterCanProcessStatusReports() got = %+v, expected %+v", res, tt.want)
			}

		})
	}
}

func TestNewDataPlaneClusterService_GetDataPlaneClusterConfig(t *testing.T) {
	type fields struct {
		clusterService ClusterService
		config         *config.ApplicationConfig
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		want    *api.DataPlaneClusterConfig
	}{
		{
			name: "should success",
			fields: fields{
				clusterService: &ClusterServiceMock{
					FindClusterByIDFunc: func(clusterID string) (*api.Cluster, *errors.ServiceError) {
						return &api.Cluster{}, nil
					},
				},
				config: &config.ApplicationConfig{
					ObservabilityConfiguration: &config.ObservabilityConfiguration{
						ObservabilityConfigRepo:        "test-repo",
						ObservabilityConfigChannel:     "test-channel",
						ObservabilityConfigAccessToken: "test-token",
						ObservabilityConfigTag:         "test-tag",
					},
				},
			},
			wantErr: false,
			want: &api.DataPlaneClusterConfig{Observability: api.DataPlaneClusterConfigObservability{
				AccessToken: "test-token",
				Channel:     "test-channel",
				Repository:  "test-repo",
				Tag:         "test-tag",
			}},
		},
		{
			name: "should fail",
			fields: fields{
				clusterService: &ClusterServiceMock{
					FindClusterByIDFunc: func(clusterID string) (*api.Cluster, *errors.ServiceError) {
						return nil, errors.NotFound("not found")
					},
				},
				config: &config.ApplicationConfig{
					ObservabilityConfiguration: &config.ObservabilityConfiguration{
						ObservabilityConfigRepo:        "test-repo",
						ObservabilityConfigChannel:     "test-channel",
						ObservabilityConfigAccessToken: "test-token",
						ObservabilityConfigTag:         "test-tag",
					},
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDataPlaneClusterService(tt.fields.clusterService, nil, tt.fields.config)
			config, err := s.GetDataPlaneClusterConfig(context.TODO(), "test-cluster-id")
			if err != nil && !tt.wantErr {
				t.Fatalf("unexpected error %v", err)
			}
			if !reflect.DeepEqual(config, tt.want) {
				t.Fatalf("result doesn't match. want: %v got: %v", tt.want, config)
			}
		})
	}
}

func Test_DataPlaneCluster_setClusterStatus(t *testing.T) {
	type input struct {
		status                  *api.DataPlaneClusterStatus
		cluster                 *api.Cluster
		dataPlaneClusterService *dataPlaneClusterService
	}
	cases := []struct {
		name         string
		inputFactory func() (*input, *api.ClusterStatus)
		wantErr      bool
		want         api.ClusterStatus
	}{
		{
			name: "when there is capacity remaining and cluster is not ready then it is set as ready",
			inputFactory: func() (*input, *api.ClusterStatus) {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				applicationConfig := sampleValidConfig()
				kafkaConfig := applicationConfig.Kafka
				testStatus.NodeInfo.Current = 3
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.Remaining.Connections = kafkaConfig.KafkaCapacity.TotalMaxConnections + 1
				testStatus.Remaining.Partitions = kafkaConfig.KafkaCapacity.MaxPartitions + 1
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterFull,
				}
				ocmClient := &ocm.ClientMock{}
				var spyReceivedUpdateStatus *api.ClusterStatus = new(api.ClusterStatus)

				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
					UpdateStatusFunc: func(cluster api.Cluster, status api.ClusterStatus) error {
						if cluster.ClusterID != apiCluster.ClusterID {
							return errors.GeneralError("unexpected test error")
						}
						*spyReceivedUpdateStatus = status
						return nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, applicationConfig)
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}, spyReceivedUpdateStatus
			},
			want:    api.ClusterReady,
			wantErr: false,
		},

		{
			name: "when there is no capacity remaining and current number of nodes is less than restricted ceiling then state is set as scaling in progress",
			inputFactory: func() (*input, *api.ClusterStatus) {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				applicationConfig := sampleValidConfig()
				testStatus.NodeInfo.Current = 3
				testStatus.NodeInfo.Ceiling = 10000
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.Remaining.Connections = 0
				testStatus.Remaining.Partitions = 0
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				var spyReceivedUpdateStatus *api.ClusterStatus = new(api.ClusterStatus)

				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
					UpdateStatusFunc: func(cluster api.Cluster, status api.ClusterStatus) error {
						if cluster.ClusterID != apiCluster.ClusterID {
							return errors.GeneralError("unexpected test error")
						}
						*spyReceivedUpdateStatus = status
						return nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, applicationConfig)
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}, spyReceivedUpdateStatus
			},
			want:    api.ClusterComputeNodeScalingUp,
			wantErr: false,
		},
		{
			name: "when there is no capacity remaining and current number of nodes is higher or equal than restricted ceiling then state is set to full",
			inputFactory: func() (*input, *api.ClusterStatus) {
				testStatus := sampleValidBaseDataPlaneClusterStatusRequest()
				applicationConfig := sampleValidConfig()
				testStatus.NodeInfo.Current = 10
				testStatus.NodeInfo.Ceiling = 11
				testStatus.NodeInfo.CurrentWorkLoadMinimum = 3
				testStatus.Remaining.Connections = 0
				testStatus.Remaining.Partitions = 0
				apiCluster := &api.Cluster{
					ClusterID: testClusterID,
					MultiAZ:   true,
					Status:    api.ClusterReady,
				}
				ocmClient := &ocm.ClientMock{}
				var spyReceivedUpdateStatus *api.ClusterStatus = new(api.ClusterStatus)

				clusterService := &ClusterServiceMock{
					SetComputeNodesFunc: func(clusterID string, numNodes int) (*v1.Cluster, *errors.ServiceError) {
						if clusterID != apiCluster.ClusterID {
							return nil, errors.GeneralError("unexpected test error")
						}
						return nil, nil
					},
					UpdateStatusFunc: func(cluster api.Cluster, status api.ClusterStatus) error {
						if cluster.ClusterID != apiCluster.ClusterID {
							return errors.GeneralError("unexpected test error")
						}
						*spyReceivedUpdateStatus = status
						return nil
					},
				}
				dataPlaneClusterService := NewDataPlaneClusterService(clusterService, ocmClient, applicationConfig)
				return &input{
					status:                  testStatus,
					cluster:                 apiCluster,
					dataPlaneClusterService: dataPlaneClusterService,
				}, spyReceivedUpdateStatus
			},
			want:    api.ClusterFull,
			wantErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f, spyReceivedStatus := tt.inputFactory()
			if f == nil {
				t.Fatalf("dataPlaneClusterService is nil")
			}

			res := f.dataPlaneClusterService.setClusterStatus(f.cluster, f.status)
			if !reflect.DeepEqual(res != nil, tt.wantErr) {
				t.Errorf("setClusterStatus() got = %+v, expected %+v", res, tt.wantErr)
			}
			if spyReceivedStatus == nil {
				t.Fatalf("spyStatus is nil")
			}
			if !reflect.DeepEqual(*spyReceivedStatus, tt.want) {
				t.Errorf("setClusterStatus() got = %+v, expected %+v", spyReceivedStatus, tt.want)
			}
		})
	}
}

func sampleValidBaseDataPlaneClusterStatusRequest() *api.DataPlaneClusterStatus {
	return &api.DataPlaneClusterStatus{
		Conditions: []api.DataPlaneClusterStatusCondition{
			api.DataPlaneClusterStatusCondition{
				Type:   "Ready",
				Status: "True",
			},
		},
		NodeInfo: api.DataPlaneClusterStatusNodeInfo{
			Ceiling:                0,
			Floor:                  0,
			Current:                0,
			CurrentWorkLoadMinimum: 0,
		},
		Remaining: api.DataPlaneClusterStatusCapacity{
			Connections:                   0,
			Partitions:                    0,
			IngressEgressThroughputPerSec: "",
			DataRetentionSize:             "",
		},
		ResizeInfo: api.DataPlaneClusterStatusResizeInfo{
			NodeDelta: multiAZClusterNodeScalingMultiple,
			Delta: api.DataPlaneClusterStatusCapacity{
				Connections:                   0,
				Partitions:                    0,
				IngressEgressThroughputPerSec: "",
				DataRetentionSize:             "",
			},
		},
	}
}

func sampleValidConfig() *config.ApplicationConfig {
	return &config.ApplicationConfig{Kafka: &config.KafkaConfig{
		KafkaCapacity: config.KafkaCapacityConfig{
			MaxPartitions:       100,
			TotalMaxConnections: 100,
		},
	}}
}