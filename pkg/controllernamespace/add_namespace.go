package controllernamespace

import (
	"github.com/openshift/openshift-azure/pkg/controllernamespace/namespace"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, namespace.Add)
}
