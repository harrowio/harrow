app = angular.module("harrowApp")

InvitationShowCtrl = (
  @invitation
  @flash
  @$state
  @$translate
) ->
  @goToProject() if @invitation.isAccepted()
  @$state.go("dashboard") if @invitation.isRefused()
  @

InvitationShowCtrl::accept = () ->
  @invitation.accept().then () =>
    @flash.success = @$translate.instant("invitations.show.flashes.accept.success")
    @goToProject()
    return
  .catch () =>
    @flash.error = @$translate.instant("invitations.show.flashes.accept.fail")
    return

InvitationShowCtrl::refuse = () ->
  @invitation.refuse().then () =>
    @flash.success = @$translate.instant("invitations.show.flashes.refuse.success")
    @$state.go("dashboard")
    return
  .catch () =>
    @flash.error = @$translate.instant("invitations.show.flashes.refuse.fail")
    return

InvitationShowCtrl::goToProject = () ->
  @$state.go("projects/show", {projectUuid: @invitation.subject.projectUuid})

app.controller("invitationShowCtrl", InvitationShowCtrl)
