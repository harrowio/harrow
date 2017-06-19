package limits

type DummyBillingHistory struct {
	planUuidFor   map[string]string
	extraProjects map[string]int
	extraUsers    map[string]int
}

func NewDummyBillingHistory() *DummyBillingHistory {
	return &DummyBillingHistory{
		planUuidFor:   map[string]string{},
		extraProjects: map[string]int{},
		extraUsers:    map[string]int{},
	}
}

func (self *DummyBillingHistory) Add(organizationUuid, billingPlanUuid string) *DummyBillingHistory {
	self.planUuidFor[organizationUuid] = billingPlanUuid
	return self
}

func (self *DummyBillingHistory) SetExtraProjects(organizationUuid string, extraProjects int) *DummyBillingHistory {
	self.extraProjects[organizationUuid] = extraProjects
	return self
}

func (self *DummyBillingHistory) SetExtraUsers(organizationUuid string, extraUsers int) *DummyBillingHistory {
	self.extraUsers[organizationUuid] = extraUsers
	return self
}

func (self *DummyBillingHistory) PlanUuidFor(organizationUuid string) string {
	return self.planUuidFor[organizationUuid]
}

func (self *DummyBillingHistory) ExtraProjectsFor(organizationUuid string) int {
	return self.extraProjects[organizationUuid]
}

func (self *DummyBillingHistory) ExtraUsersFor(organizationUuid string) int {
	return self.extraUsers[organizationUuid]
}
