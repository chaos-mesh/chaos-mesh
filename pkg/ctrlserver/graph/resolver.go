package graph

//go:generate gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver struct {
	Log       logr.Logger
	Client    client.Client
	Clientset kubernetes.Interface
}
