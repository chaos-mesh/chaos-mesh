// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

type Color string

const (
	Blue    Color = "Blue"
	Red     Color = "Red"
	Green   Color = "Green"
	Cyan    Color = "Cyan"
	NoColor Color = ""
)

var (
	colorFunc = map[Color]func(string, ...interface{}){
		Blue:  color.Blue,
		Red:   color.Red,
		Green: color.Green,
		Cyan:  color.Cyan,
	}
	scheme = runtime.NewScheme()
)

// ClientSet contains two different clients
type ClientSet struct {
	CtrlCli client.Client
	KubeCli *kubernetes.Clientset
}

type ChaosResult struct {
	Name string
	Pods []PodResult
}

type PodResult struct {
	Name  string
	Items []ItemResult
}

const (
	ItemSuccess = iota + 1
	ItemFailure
)

const ChaosDaemonClientCert = "chaos-mesh-daemon-client-certs"
const ChaosDaemonNamespace = "chaos-testing"

var TLSFiles grpcUtils.TLSFile
var Insecure bool

type ItemResult struct {
	Name    string
	Value   string
	Status  int    `json:",omitempty"`
	SucInfo string `json:",omitempty"`
	ErrInfo string `json:",omitempty"`
}

func init() {
	_ = v1alpha1.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)
}

func upperCaseChaos(str string) string {
	parts := regexp.MustCompile("(.*)(chaos)").FindStringSubmatch(str)
	return strings.Title(parts[1]) + strings.Title(parts[2])
}

// PrettyPrint print with tab number and color
func PrettyPrint(s string, indentLevel int, color Color) {
	var tabStr string
	for i := 0; i < indentLevel; i++ {
		tabStr += "\t"
	}
	str := fmt.Sprintf("%s%s\n\n", tabStr, regexp.MustCompile("\n").ReplaceAllString(s, "\n"+tabStr))
	if color != NoColor {
		if cfunc, ok := colorFunc[color]; !ok {
			fmt.Print("COLOR NOT SUPPORTED")
		} else {
			cfunc(str)
		}
	} else {
		fmt.Print(str)
	}
}

// PrintResult prints result to users in prettier format
func PrintResult(result []*ChaosResult) {
	for _, chaos := range result {
		PrettyPrint("[Chaos]: "+chaos.Name, 0, Blue)
		for _, pod := range chaos.Pods {
			PrettyPrint("[Pod]: "+pod.Name, 0, Blue)
			for i, item := range pod.Items {
				PrettyPrint(fmt.Sprintf("%d. [%s]", i+1, item.Name), 1, Cyan)
				PrettyPrint(item.Value, 1, NoColor)
				if item.Status == ItemSuccess {
					if item.SucInfo != "" {
						PrettyPrint(item.SucInfo, 1, Green)
					} else {
						PrettyPrint("Execute as expected", 1, Green)
					}
				} else if item.Status == ItemFailure {
					PrettyPrint(fmt.Sprintf("Failed: %s ", item.ErrInfo), 1, Red)
				}
			}
		}
	}
}

// MarshalChaos returns json in readable format
func MarshalChaos(s interface{}) (string, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal indent")
	}
	return string(b), nil
}

// InitClientSet inits two different clients that would be used
func InitClientSet() (*ClientSet, error) {
	restconfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	ctrlClient, err := client.New(restconfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client")
	}
	kubeClient, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return nil, errors.Wrap(err, "error in getting acess to k8s")
	}
	return &ClientSet{ctrlClient, kubeClient}, nil
}

// Log print log of pod
func Log(pod v1.Pod, tail int64, c *kubernetes.Clientset) (string, error) {
	podLogOpts := v1.PodLogOptions{}
	//use negative tail to indicate no tail limit is needed
	if tail >= 0 {
		podLogOpts.TailLines = func(i int64) *int64 { return &i }(tail)
	}

	req := c.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	// FIXME: get context from parameter
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", errors.Wrapf(err, "failed to open log stream for pod %s/%s", pod.GetNamespace(), pod.GetName())
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to copy information from podLogs to buf")
	}
	return buf.String(), nil
}
