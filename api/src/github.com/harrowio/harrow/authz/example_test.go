package authz_test

import (
	"fmt"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/config"
)

type Foo struct{}

// AuthorizationName supplies the name of the authz.Subject for use
// in capablities.
func (self *Foo) AuthorizationName() string { return "foo" }

type FooOwner struct{}

// Capabilities implements authz.Role by returning the list of
// capabilities for this role.
func (self *FooOwner) Capabilities() []string {
	return []string{"delete-foo", "update-foo"}
}

// This example shows how to make an object compatible with authz.
func Example_minimalImplementation() {
	role := &FooOwner{}
	service := authz.NewService(nil, nil, config.GetConfig())

	// authz doesn't know anything about FooOwner
	service.AddRole(role)

	subject := &Foo{}

	allowed, err := service.Can("update", subject)
	if err != nil {
		fmt.Printf("service.Can(%q, %#v): %s\n", "update", subject, err)
	}

	if !allowed {
		fmt.Printf("%T cannot update %T\n", role, subject)
	} else {
		fmt.Printf("%T can update %T\n", role, subject)
	}

	// Output:
	// *authz_test.FooOwner can update *authz_test.Foo
}
