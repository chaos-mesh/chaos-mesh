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

package types

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Controller string

// Object only used for registration webhook for various Kind of chaos custom resources.
// Deprecated: use WebhookObject instead.
// TODO: migrate it to WebhookObject
type Object struct {
	// Object should be the same as the kind of the chaos custom resource.
	Object v1alpha1.InnerObject
	// Name indicates the name of the webhook. It would be used to dedicate enabling the webhook for this Kind of
	// chaos custom resource or not.
	Name string
}

// ChaosObjects is the list of all kind of chaos custom resource, following the registration pattern.
// Deprecated: use WebhookObjects instead.
var ChaosObjects = fx.Supply(
	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "awschaos",
			Object: &v1alpha1.AWSChaos{},
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
			Object: &v1alpha1.GCPChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "physicalmachinechaos",
			Object: &v1alpha1.PhysicalMachineChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "azurechaos",
			Object: &v1alpha1.AzureChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "blockchaos",
			Object: &v1alpha1.BlockChaos{},
		},
	},

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "ciliumchaos",
			Object: &v1alpha1.CiliumChaos{},
		},
	},
)

// WebhookObject only used for registration the
type WebhookObject struct {
	// Object should be the same as the kind of the chaos custom resource.
	Object v1alpha1.WebhookObject
	// Name indicates the name of the webhook. It would be used to dedicate enabling the webhook for this Kind of
	// chaos custom resource or not.
	Name string
}

// WebhookObjects is the list of all kind of chaos custom resource, following the registration pattern.
// When you add a new kind of chaos custom resource, please add it to the list.
var WebhookObjects = fx.Supply(
	fx.Annotated{
		Group: "webhookObjs",
		Target: WebhookObject{
			Name:   "physicalmachine",
			Object: &v1alpha1.PhysicalMachine{},
		},
	},
	fx.Annotated{
		Group: "webhookObjs",
		Target: WebhookObject{
			Name:   "statuscheck",
			Object: &v1alpha1.StatusCheck{},
		},
	},
	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "cloudstackvmchaos",
			Object: &v1alpha1.CloudStackVMChaos{},
		},
	},
)
