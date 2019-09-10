package podchaos

import (
	"time"

	"github.com/cwen0/chaos-operator/pkg/client/clientset/versioned"
	informers "github.com/cwen0/chaos-operator/pkg/client/informers/externalversions"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// Controller is the controller implementation for pod chaos resources.
type Controller struct {
}

// NewController returns a new pod chaos controller.
func NewController(
	_ kubernetes.Interface,
	_ versioned.Interface,
	_ kubeinformers.SharedInformerFactory,
	_ informers.SharedInformerFactory,
) *Controller {
	return &Controller{}
}

// Run runs the podchaos controller.
func (c *Controller) Run() {
	for {
		time.Sleep(time.Second * 3)
	}
}
