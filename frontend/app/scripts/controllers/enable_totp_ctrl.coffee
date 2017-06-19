app = angular.module("harrowApp")

EnableTotpCtrl = (
  @$state
  @$translate
  @flash
  @authentication
  @totpSecret
) ->
  @totpToken = ''
  @totpEmail = @authentication.currentUser.subject.email

  @

EnableTotpCtrl::enable = () ->
  @authentication.currentUser.enableTotp(@totpToken).then =>
    @$state.go("settings.mfa")
    return
  .catch (response) =>
    @flash.error = @$translate.instant("forms.enableTotp.flashes.confirmationFailed")


app.controller("enableTotpCtrl", EnableTotpCtrl)
