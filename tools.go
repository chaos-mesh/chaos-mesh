// +build tools

package tools

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "github.com/mgechev/revive"
	_ "github.com/pingcap/failpoint/failpoint-ctl"
	_ "golang.org/x/tools/cmd/goimports"
)