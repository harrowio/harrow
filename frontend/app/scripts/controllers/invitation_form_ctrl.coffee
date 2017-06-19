app = angular.module("harrowApp")

InvitationFormCtrl = (
  @invitation
  @organization
  @project
  @flash
  @$state
  @$translate
  @invitationResource
  @$q
) ->
  @

InvitationFormCtrl::save = () ->
  @invitationResource.save(@invitation).then (invitation) =>
    @flash.success = @$translate.instant("forms.invitationForm.flashes.success", invitation.subject)
    @$state.go("projects/edit.people", {projectUuid: @project.subject.uuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.invitationForm.flashes.fail", @invitation.subject)
    @$q.reject(reason)

app.controller("invitationFormCtrl", InvitationFormCtrl)
