// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Controller string

type Object struct {
	Object v1alpha1.InnerObject
	Name   string
}

var ChaosObjects = fx.Supply(
	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "awschaos",
			Object: &v1alpha1.AwsChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "dnschaos",
			Object: &v1alpha1.DNSChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "httpchaos",
			Object: &v1alpha1.HTTPChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "iochaos",
			Object: &v1alpha1.IOChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "kernelchaos",
			Object: &v1alpha1.KernelChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "jvmchaos",
			Object: &v1alpha1.JVMChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "networkchaos",
			Object: &v1alpha1.NetworkChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "podchaos",
			Object: &v1alpha1.PodChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "stresschaos",
			Object: &v1alpha1.StressChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "timechaos",
			Object: &v1alpha1.TimeChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "gcpchaos",
			Object: &v1alpha1.GcpChaos{},
		},
	},
)
