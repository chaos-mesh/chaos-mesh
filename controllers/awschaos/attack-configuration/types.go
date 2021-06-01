package attack_configuration

import (
	"context"
	"errors"
	"fmt"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/go-logr/logr"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
)

type endpoint struct {
	ctx.Context

}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	panic("implement me")
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	panic("implement me")
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AwsChaos)
	if !ok {
		err := errors.New("chaos is not awschaos")
		e.Log.Error(err, "chaos is not AwsChaos", "chaos", chaos)
		return err
	}

	pod := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Name: 	fmt.Sprintf("job%s", req.NamespacedName),
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			RestartPolicy: "Never",
			Containers: []v1.Container{
				{
					Name: "attackConfiguration",
					Image: "",
					Args: { },
				},
			},
		},
	}
	e.Log.Info("Running the Container")

	return nil

}
func init() {
	router.Register("awschaos", &v1alpha1.AwsChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.AwsChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.AwsssmChaos
	},
	func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
