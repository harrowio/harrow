app = angular.module("harrowApp")

ProjectMemberListCtrl = (
  @organization
  @project
  @members
  @authentication
  @projectMemberResource
  @$translate
  @$state
  @$scope
  @flash
) ->
  @user = @authentication.currentUser
  @userHierarchyLevel = @membershipTypeHierarchyLevel(@userMembership())
  @

ProjectMemberListCtrl::userMembership = () ->
  userMembers = @members.filter (member) =>
    member.subject.uuid == @user.subject.uuid
  return userMembers[0].subject.type if userMembers.length > 0
  return 'guest'

ProjectMemberListCtrl::membershipTypeHierarchyLevel = (membershipType) ->
  switch membershipType
    when 'owner' then 40
    when 'manager' then 30
    when 'member' then 20
    when 'guest' then 10
    else 0

ProjectMemberListCtrl::canPromote = (member) ->
   @userHierarchyLevel > @membershipTypeHierarchyLevel(member.subject.type)

ProjectMemberListCtrl::remove = (toRemove) ->
   return unless @confirmRemove(toRemove)
   toRemove.remove(@project.subject.uuid).then () =>
     @members = @members.filter (current) ->
       current.subject.uuid != toRemove.subject.uuid

ProjectMemberListCtrl::confirmRemove = (member) ->
   confirm(@$translate.instant("projectMemberships.prompts.remove", {
           name: member.subject.name,
   }))

ProjectMemberListCtrl::promote = (member) ->
   newType = @newMembershipType(member.subject.type)

   return unless @confirmPromote(member, newType)

   @projectMemberResource.update(member).then (updatedMember) =>
     angular.forEach @members, (member, index) =>
       if member.subject.uuid == updatedMember.subject.uuid
         @members[index] = updatedMember

ProjectMemberListCtrl::newMembershipType = (membershipType) ->
  switch membershipType
    when 'owner' then 'owner'
    when 'manager' then 'owner'
    when 'member' then 'manager'
    when 'guest' then 'member'
    else 'guest'

ProjectMemberListCtrl::confirmPromote = (member, newType) ->
  confirm(@$translate.instant("projectMemberships.prompts.promote", {
          name: member.subject.name,
          type: newType,
  }))

app.controller("projectMemberListCtrl", ProjectMemberListCtrl)
