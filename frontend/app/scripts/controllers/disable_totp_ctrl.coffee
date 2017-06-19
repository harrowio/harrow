app = angular.module("harrowApp")

DisableTotpCtrl = (
  @$state
  @$translate
  @flash
  @authentication
  @userResource
) ->
  @

DisableTotpCtrl::confirm = () ->
  totp = @session.subject.totp
  @authentication.currentUser.disableTotp(totp).then () =>
    @authentication.reloadCurrentUser().then () =>
      @$state.go("settings.mfa")
      return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.disableTotp.flashes.confirmationFailed")

app.controller("disableTotpCtrl", DisableTotpCtrl)
