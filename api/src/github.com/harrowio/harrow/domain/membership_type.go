package domain

const (
	guest   string = "guest"
	member  string = "member"
	manager string = "manager"
	owner   string = "owner"
)

var MembershipTypeGuest string = guest
var MembershipTypeMember string = member
var MembershipTypeManager string = manager
var MembershipTypeOwner string = owner

func MembershipTypeHierarchyLevel(membershipType string) int {
	switch membershipType {
	case guest:
		return 10
	case member:
		return 20
	case manager:
		return 30
	case owner:
		return 40
	default:
		return 0
	}
}
