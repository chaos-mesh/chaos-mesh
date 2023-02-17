package physicalmachine

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func NewPhysicalMachineTemplateCmd() (*cobra.Command, error) {
	templateCmd := &cobra.Command{
		Use:   `template`,
		Short: `Generate deploy topology.yaml`,
		Long: `Generate deploy topology.yaml
				Examples:
				# Generate deploy topology.yaml
				chaosctl pm template >  topology.yaml
				`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			to := &Topology{
				NamePrefix:         "chaosd",
				ChaosMeshNamespace: "chaos-mesh",
				ChaosdPort:         31768,
				SshPort:            22,
				CAPath:             "/etc/chaosd/pki",
				SrcPath:            "/etc/chaosd/bin",
				DstPath:            "/etc/chaosd/bin",
				FielName:           "chaosd.tar.gz",
				Namespace:          "default",
				Labels:             map[string]string{"app": "name"},
				RemoteIPs:          []string{"192.168.1.1"},
			}
			result, err := yaml.Marshal(to)
			if err != nil {
				return err
			}
			fmt.Println(string(result))
			return nil
		},
	}
	return templateCmd, nil
}
