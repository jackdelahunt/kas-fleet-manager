package cloudprovider

import (
	"encoding/json"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/api/public"
	presenters2 "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/presenters"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/internal/kafka/internal/services"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/environments"
	"github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/flags"
	coreServices "github.com/bf2fc6cc711aee1a0c2a/kas-fleet-manager/pkg/services"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

// NewRegionsListCommand creates a new command for listing regions.
func NewRegionsListCommand(env *environments.Env) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regions",
		Short: "lists all supported cloud providers",
		Long:  "lists all supported cloud providers",
		Run: func(cmd *cobra.Command, args []string) {
			runRegionsList(env, cmd, args)
		},
	}
	cmd.Flags().String(FlagID, "aws", "Cloud provider id")
	return cmd
}

func runRegionsList(env *environments.Env, cmd *cobra.Command, _ []string) {

	id := flags.MustGetDefinedString(FlagID, cmd.Flags())

	var config coreServices.ConfigService
	var cloudProviderService services.CloudProvidersService
	env.MustResolveAll(&config, &cloudProviderService)

	cloudRegions, err := cloudProviderService.ListCloudProviderRegions(id)
	if err != nil {
		glog.Fatalf("Unable to list cloud provider regions: %s", err.Error())
	}

	regionList := public.CloudRegionList{
		Kind:  "CloudRegionList",
		Total: int32(len(cloudRegions)),
		Size:  int32(len(cloudRegions)),
		Page:  int32(1),
	}
	for _, cloudRegion := range cloudRegions {
		cloudRegion.Enabled = config.IsRegionSupportedForProvider(cloudRegion.CloudProvider, cloudRegion.Id)
		converted := presenters2.PresentCloudRegion(&cloudRegion)
		regionList.Items = append(regionList.Items, converted)
	}

	output, marshalErr := json.MarshalIndent(regionList, "", "    ")
	if marshalErr != nil {
		glog.Fatalf("Failed to format  cloud provider region list: %s", err.Error())
	}

	glog.V(10).Infof("%s", output)

}