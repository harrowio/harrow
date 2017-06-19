package domain

import "fmt"

// capabilityList is a helper for constructing lists of capabilities.
type capabilityList []string

func newCapabilityList() *capabilityList {
	capabilities := capabilityList([]string{})
	return &capabilities
}

func (l *capabilityList) writesFor(thing string) *capabilityList {
	*l = append(*l,
		"create-"+thing,
		"update-"+thing,
		"archive-"+thing,
	)

	return l
}

func (l *capabilityList) reads(things ...string) *capabilityList {
	for _, thing := range things {
		*l = append(*l, "read-"+thing)
	}

	return l
}

func (l *capabilityList) creates(things ...string) *capabilityList {
	for _, thing := range things {
		*l = append(*l, "create-"+thing)
	}

	return l
}

func (l *capabilityList) updates(things ...string) *capabilityList {
	for _, thing := range things {
		*l = append(*l, "update-"+thing)
	}

	return l
}

func (l *capabilityList) archives(things ...string) *capabilityList {
	for _, thing := range things {
		*l = append(*l, "archive-"+thing)
	}

	return l
}

func (l *capabilityList) does(action string, things ...string) *capabilityList {
	for _, thing := range things {
		*l = append(*l, fmt.Sprintf("%s-%s", action, thing))
	}

	return l
}

func (l *capabilityList) add(capabilities []string) *capabilityList {
	set := map[string]bool{}
	for _, cap := range *l {
		set[cap] = true
	}

	for _, cap := range capabilities {
		set[cap] = true
	}

	merged := make([]string, 0, len(set))
	for cap, _ := range set {
		merged = append(merged, cap)
	}

	result := capabilityList(merged)
	return &result
}

func (l *capabilityList) strings() []string {
	return *l
}

func (l *capabilityList) Capabilities() []string {
	return l.strings()
}
