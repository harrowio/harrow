app = angular.module("harrowApp")

SessionConfirmationCtrl = (
  @$state
  @$stateParams
  @$translate
  @flash
  @authentication
) ->
  @invitationUuid = @$stateParams.invitation
  @origin = @$stateParams.origin
  @

SessionConfirmationCtrl::confirm = () ->
  @authentication.confirm(@session.subject.totp).then () =>
    if @invitationUuid
        @$state.go("invitations/show", {uuid: @invitationUuid})
    else
      if @origin
        window.location.hash = @origin
      else
        @$state.go("dashboard")
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.sessionConfirmation.flashes.confirmationFailed")
    return

app.controller("sessionConfirmationCtrl", SessionConfirmationCtrl)
