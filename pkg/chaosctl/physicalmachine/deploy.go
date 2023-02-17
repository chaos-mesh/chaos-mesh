package physicalmachine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type PhysicalMachineDeployOptions struct {
	sshUser           string
	sshPassword       bool
	sshPrivateKeyFile string
	sshKnowHost       bool
}

func NewPhysicalMachineDeployCmd() (*cobra.Command, error) {

	deployOption := &PhysicalMachineDeployOptions{}

	deployCmd := &cobra.Command{
		Use:   `deploy (topology.yaml)`,
		Short: `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster, and deploy chaosd in physical machine`,
		Long: `Generate TLS certs for certain physical machine automatically, and create PhysicalMachine CustomResource in Kubernetes cluster, and deploy chaosd in physical machine

Examples:
  # Generate TLS certs for remote physical machine, create PhysicalMachine CustomResource in certain namespace, and deploy chaosd in physical machine 
  chaosctl pm deploy /path/topology.yaml
  `,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deployOption.Validate(); err != nil {
				return err
			}
			return deployOption.Run(args)
		},
	}
	deployCmd.PersistentFlags().StringVarP(&deployOption.sshUser, "user", "u", "root", "username for ssh connection")
	deployCmd.PersistentFlags().StringVarP(&deployOption.sshPrivateKeyFile, "key", "k", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"), "private key filepath for ssh connection")
	deployCmd.PersistentFlags().BoolVarP(&deployOption.sshPassword, "password", "p", false, "password for ssh connection")
	deployCmd.PersistentFlags().BoolVarP(&deployOption.sshKnowHost, "knowhost", "c", false, "check know host key")
	return deployCmd, nil
}

func (o *PhysicalMachineDeployOptions) Validate() error {
	if len(o.sshUser) == 0 {
		return errors.New("--ssh-user must be specified")
	}
	return nil
}

func (o *PhysicalMachineDeployOptions) Run(args []string) error {

	if len(args) < 1 {
		return errors.New("physical machine topology.yaml is required")
	}
	to := &Topology{}
	if err := ParseTopologyYaml(args[0], to); err != nil {
		return errors.Wrap(err, "parse topology.yaml err")
	}

	if len(to.RemoteIPs) == 0 {
		return errors.New("physical machine list is empty")
	}

	sshConfig, err := getSshTunnelConfig(o.sshUser, o.sshPrivateKeyFile, o.sshPassword, o.sshKnowHost)
	if err != nil {
		return err
	}

	clientset, err := common.InitClientSet()
	if err != nil {
		return err
	}

	ctx := context.Background()
	caCert, caKey, err := GetChaosdCAFileFromCluster(ctx, to.ChaosMeshNamespace, clientset.CtrlCli)
	if err != nil {
		return err
	}

	// generate chaosd cert and private key
	serverCert, serverKey, err := NewCertAndKey(caCert, caKey)
	if err != nil {
		return err
	}

	for _, ip := range to.RemoteIPs {
		physicalMachineName := formatPhysicalMachineName(to.NamePrefix, ip)
		address := formatAddress(ip, to.ChaosdPort, true)
		sshTunnel, err := NewSshTunnel(ip, strconv.Itoa(to.SshPort), sshConfig)
		if err != nil {
			return err
		}
		if err := sshTunnel.Open(); err != nil {
			return err
		}
		defer sshTunnel.Close()

		if err := writeCertAndKeyToRemote(sshTunnel, to.CAPath, ChaosdPkiName, serverCert, serverKey); err != nil {
			return err
		}
		if err := writeCertToRemote(sshTunnel, to.CAPath, "ca", caCert); err != nil {
			return err
		}
		common.PrettyPrint(fmt.Sprintf("%s write cert to remote success", ip), 0, common.Cyan)

		if err := writeChaosdToRemote(sshTunnel, to.SrcPath, to.DstPath, to.FielName); err != nil {
			return err
		}
		common.PrettyPrint(fmt.Sprintf("%s write chaosd to remote success", ip), 0, common.Cyan)

		if err := deployChaosd(sshTunnel, to); err != nil {
			return err
		}
		common.PrettyPrint(fmt.Sprintf("%s deploy chaosd success", ip), 0, common.Cyan)

		if err := CreatePhysicalMachine(ctx, clientset.CtrlCli, to.Namespace, physicalMachineName, address, to.Labels); err != nil {
			return err
		}
		common.PrettyPrint(fmt.Sprintf("%s create PhysicalMachine success", ip), 0, common.Cyan)
	}
	return nil

}
func deployChaosd(sshTunnel *SshTunnel, to *Topology) error {
	var (
		dstFilePath = pathForChaosd(to.DstPath, to.FielName)
		binPath     = pathForChaosd(to.DstPath, ChaosdPkiName)
		caPath      = pathForCert(to.CAPath, "ca")
		certPath    = pathForCert(to.CAPath, ChaosdPkiName)
		keyPath     = pathForKey(to.CAPath, ChaosdPkiName)
	)
	sshTunnel.Exec(fmt.Sprintf("tar zxvf %s -C %s --strip-components 1", dstFilePath, to.DstPath))

	cmd := fmt.Sprintf(" %s server --https-port %d --CA=%s --cert=%s --key=%s > /var/log/chaosd.log 2>&1 & ", binPath, to.ChaosdPort, caPath, certPath, keyPath)
	return sshTunnel.Exec(cmd)

}

func writeChaosdToRemote(sshTunnel *SshTunnel, src, dst, fileName string) error {
	srcFile, err := os.Open(pathForChaosd(src, fileName))
	if err != nil {
		return errors.Wrapf(err, "open file %s err", src)
	}
	return sshTunnel.SFTP(pathForChaosd(dst, fileName), srcFile)
}

func pathForChaosd(path, fileName string) string {
	return filepath.Join(path, fileName)
}

func ParseTopologyYaml(file string, out *Topology) error {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err = yaml.UnmarshalStrict(yamlFile, out); err != nil {
		return err
	}
	return nil
}

func formatPhysicalMachineName(prefix, ip string) string {
	return fmt.Sprintf("%s-%s", prefix, ip)
}

type Topology struct {
	NamePrefix         string            `yaml:"name_preix,omitempty" default:"chaosd-"`
	ChaosMeshNamespace string            `yaml:"chaos_mesh_namespace,omitempty" default:"chaos-mesh"`
	ChaosdPort         int               `yaml:"chaosd_port,omitempty" default:"31768"`
	SshPort            int               `yaml:"ssh_port,omitempty" default:"22"`
	CAPath             string            `yaml:"ca_path,omitempty" default:"/etc/chaosd/pki"`
	SrcPath            string            `yaml:"src_path,omitempty" default:"/etc/chaosd/bin"`
	DstPath            string            `yaml:"dst_path,omitempty" default:"/etc/chaosd/bin"`
	FielName           string            `yaml:"file_name,omitempty" default:"chaosd.tar.gz"`
	Namespace          string            `yaml:"namespace,omitempty" default:"default"`
	Labels             map[string]string `yaml:"labels,omitempty"`
	RemoteIPs          []string          `yaml:"remote_ips,omitempty"`
}
