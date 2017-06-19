/*
Package authz handles authorization logic for different entities in
Harrow.  Its purpose is to answer questions such as "Can this user edit
that project?".

Basics

This package implements role-based authorization.  The following terms
related to role-based authorization are used throughout this text:

	Action: something a user can do ("create")
	Subject:  the thing subjected to the user's action ("project")
	Capability: a combination of an action and a subject ("create-project")
	Role: a set of capabilities ([]string{"create-project", "create-task", ...})

A user is authorized to perform an action if she possesses the right
Capability.  The list of Capabilities a user possesses is determined by
her roles.  A user can possibly have many roles, as determined by the
authorization context.  The set of all Capabilities a user possesses in
relation to a Subject is the union of all her roles.

The context for authorization is determined by the user's relation to
the Subject.  Roles are determined by inspecting this context.

Implementation

The definitions above pose the question how a user's roles are determined.
Most of the roles in Harrow are related to membership in a project or
organization (e.g. ProjectOwner, OrganizationMember, etc).  When asking
for authorization directly on a project or organization, the role
can easily be determined by looking at existing project or organization
memberships.

Subjects that belong to a project need to expose their relation with
a project in a way that allows authz to find the associated project.
Likewise the relation to an organization, associated user and the owner
of the Subject need to be exposed.  The authz package defines several
interfaces for this purpose: Subject, BelongsToProject, BelongsToUser,
BelongsToOrganization, and Ownable.

Since not every Subject is associated with a project, organization or
user, the implemenation of these interfaces is optional.  Authz checks
whether the Subject implements any of these interfaces and loads the
respective entities for each implemented interface.

Authz first tries to load all the entities associated with the Subject
and then looks at them to determine the user's roles.  Once this is done,
the requested Capability is checked against the Capabilities provided
by all the roles.

*/
package authz
