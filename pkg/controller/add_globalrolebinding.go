package controller

import (
	"github.com/yagonobre/global-role-binding/pkg/controller/globalrolebinding"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, globalrolebinding.Add)
}
