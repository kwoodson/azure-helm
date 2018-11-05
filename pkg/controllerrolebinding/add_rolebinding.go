package controllerrolebinding

import "github.com/openshift/openshift-azure/pkg/controllerrolebinding/rolebinding"

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, rolebinding.Add)
}
