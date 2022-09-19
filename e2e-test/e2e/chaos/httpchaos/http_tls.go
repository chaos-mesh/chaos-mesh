package httpchaos

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

//go:embed keys
var content embed.FS

type TLSServerKeys struct {
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

func setupEmbedFiles(serverIP string) (TLSServerKeys, []byte) {
	c, err := content.ReadDir("keys")
	framework.ExpectNoError(err, "read key dir error")
	var key []byte
	var ca []byte
	for _, f := range c {
		if f.IsDir() {
			continue
		}
		b, err := content.ReadFile("keys/" + f.Name())
		framework.ExpectNoError(err, "read key file error")
		switch f.Name() {
		case "server.key":
			key = b
		case "ca.crt":
			ca = b
		}
		err = os.WriteFile(f.Name(), b, 0644)
		framework.ExpectNoError(err, "write key file error")
	}

	f, err := os.OpenFile("server.ext", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		framework.ExpectNoError(err, "open server.ext file error")
	}
	if _, err = f.WriteString(fmt.Sprint("IP.1 = " + serverIP)); err != nil {
		framework.ExpectNoError(err, "write server.ext file error")
	}
	err = f.Close()
	framework.ExpectNoError(err, "close server.ext file error")

	cmdStr := "openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650 -sha256 -extfile server.ext"
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		framework.ExpectNoError(err, "run openssl cmd error: "+string(output))
	}
	crt, err := os.ReadFile("server.crt")
	framework.ExpectNoError(err, "read server.crt file error")

	return TLSServerKeys{
		Cert: crt,
		Key:  key,
	}, ca
}

func TestcaseHttpTLSThenRecover(
	ns string,
	kubeCli kubernetes.Interface,
	cli client.Client,
	c HTTPE2EClient,
	port uint16,
	tlsPort uint16,
) {
	serverKeys, ca := setupEmbedFiles(c.IP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("waiting on e2e helper ready")
	err := util.WaitHTTPE2EHelperReady(*c.C, c.IP, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")
	By("create http delay chaos CRD objects")

	body, err := json.Marshal(serverKeys)
	framework.ExpectNoError(err, "marshal server keys error")
	err = util.SetupHTTPE2EHelperTLSConfig(*c.C, c.IP, port, tlsPort, body)
	framework.ExpectNoError(err, "setup e2e helper tls config error")
	delay := "1ms"

	_, err = kubeCli.CoreV1().Secrets(ns).Create(ctx, &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-tls",
			Namespace: ns,
		},
		Data: map[string][]byte{
			"ca.crt":     ca,
			"server.crt": serverKeys.Cert,
			"server.key": serverKeys.Key,
		},
	}, metav1.CreateOptions{})
	framework.ExpectNoError(err, "create secret error")
	caName := "ca.crt"
	httpChaos := &v1alpha1.HTTPChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.HTTPChaosSpec{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "http"},
					},
				},
				Mode: v1alpha1.OneMode,
			},
			Port:   8081,
			Target: "Request",
			PodHttpChaosActions: v1alpha1.PodHttpChaosActions{
				Delay: &delay,
			},
			TLS: &v1alpha1.PodHttpChaosTLS{
				SecretName:      "http-tls",
				SecretNamespace: ns,
				CertName:        "server.crt",
				KeyName:         "server.key",
				CAName:          &caName,
			},
		},
	}
	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "create http chaos error")

	By("waiting for HTTP pong")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		err := util.WaitHTTPE2EHelperTLSReady(*c.C, c.IP, tlsPort)
		if err != nil {
			return false, err
		}
		return true, nil
	})
	framework.ExpectNoError(err, "http chaos doesn't work as expected")
	By("apply http chaos successfully")

	By("delete chaos CRD objects")
	// delete chaos CRD
	err = cli.Delete(ctx, httpChaos)
	framework.ExpectNoError(err, "failed to delete http chaos")
}
