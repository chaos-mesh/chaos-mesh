package graph

//go:generate gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

import "sigs.k8s.io/controller-runtime/pkg/client"

type Resolver struct {
	Client client.Client
}
