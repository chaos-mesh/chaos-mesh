package awschaos

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/awschaos/action"
	"github.com/chaos-mesh/chaos-mesh/controllers/awschaos/action/ec2"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	kubeclient "sigs.k8s.io/controller-runtime/pkg/client"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	awschaosFinalizer = "awschaos.chaos-mesh.org/finalizer"
)

// endpoint is awschaos reconciler
type endpoint struct {
	ctx.Context
}

// Object would return the instance of chaos
func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.AWSChaos{}
}

func (r *endpoint) AWSConfig(ctx context.Context, awschaos *v1alpha1.AWSChaos) (aws.Config, error) {
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(awschaos.Spec.Config.Region),
	}
	if awschaos.Spec.Config.Credential != nil {
		cred := corev1.Secret{}
		if err := r.Client.Get(ctx, kubeclient.ObjectKey{
			Namespace: awschaos.Namespace,
			Name:      awschaos.Spec.Config.Credential.Name,
		}, &cred); err != nil {
			return aws.Config{}, err
		}
		accessKeyID, secretAccessKey, err := getStaticCred(&cred)
		if err != nil {
			return aws.Config{}, err
		}
		opts = append(opts, awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)))
	}
	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		r.Log.Error(err, "unable to load aws SDK config")
		return aws.Config{}, err
	}
	return cfg, nil
}

// Apply applies awschaos
func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AWSChaos)
	if !ok {
		err := errors.New("chaos is not AWSChaos")
		r.Log.Error(err, "chaos is not AWSChaos", "chaos", chaos)
		return err
	}

	awschaos.Finalizers = finalizer.InsertFinalizer(awschaos.Finalizers, awschaosFinalizer)

	cfg, err := r.AWSConfig(ctx, awschaos)
	if err != nil {
		return err
	}

	switch awschaos.Spec.Action {
	case v1alpha1.AWSActionStop:
		stopper := newStopper(cfg, awschaos.Spec.Service, awschaos.Spec.Resource, &awschaos.Spec.Selector)
		if stopper == nil {
			return fmt.Errorf("unrecognized service: %s", awschaos.Spec.Service)
		}

		if awschaos.Status.Snapshot == nil {
			snapshot, err := stopper.Snapshot(ctx)
			if err != nil {
				return err
			}
			awschaos.Status.Snapshot = snapshot
			if err := r.Status().Update(ctx, awschaos); err != nil {
				return err
			}
		}
		if err := stopper.Stop(ctx, awschaos.Status.Snapshot); err != nil {
			r.Log.Error(err, "stop resource error")
			return err
		}
	default:
	}

	return nil
}

func getStaticCred(secret *corev1.Secret) (string, string, error) {
	b64ID := secret.Data[v1alpha1.AWSAccessKeyID]
	b64Secret := secret.Data[v1alpha1.AWSSecretAccessKey]
	id, err := base64.StdEncoding.DecodeString(string(b64ID))
	if err != nil {
		return "", "", err
	}
	sec, err := base64.StdEncoding.DecodeString(string(b64Secret))
	if err != nil {
		return "", "", err
	}
	return string(id), string(sec), nil
}

func newStopper(cfg aws.Config, svc v1alpha1.AWSService, resource string, selector *v1alpha1.AWSSelector) action.Stopper {
	switch svc {
	case v1alpha1.AWSServiceEC2:
		s := ec2.NewStopper(cfg, resource, selector)
		return s
	}
	return nil
}

func newRecoverer(cfg aws.Config, svc v1alpha1.AWSService, resource string) action.Recoverer {
	switch svc {
	case v1alpha1.AWSServiceEC2:
		s := ec2.NewRecoverer(cfg, resource)
		return s
	}
	return nil
}

// Recover means the reconciler recovers the chaos action
func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AWSChaos)
	if !ok {
		err := errors.New("chaos is not AWSChaos")
		r.Log.Error(err, "chaos is not AWSChaos", "chaos", chaos)
		return err
	}
	cfg, err := r.AWSConfig(ctx, awschaos)
	if err != nil {
		return err
	}
	switch awschaos.Spec.Action {
	case v1alpha1.AWSActionStop:
		recoverer := newRecoverer(cfg, awschaos.Spec.Service, awschaos.Spec.Resource)
		if recoverer == nil {
			return fmt.Errorf("unrecognized service: %s", awschaos.Spec.Service)
		}
		if err := recoverer.Recover(ctx, awschaos.Status.Snapshot); err != nil {
			r.Log.Error(err, "recover resource error")
			return err
		}
	default:
	}

	awschaos.Finalizers = finalizer.RemoveFromFinalizer(awschaos.Finalizers, awschaosFinalizer)

	return nil
}

func init() {
	router.Register("awschaos", &v1alpha1.AWSChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
