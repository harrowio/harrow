app = angular.module("harrowApp")

ForgotPasswordCtrl = (
  @endpoint
  @flash
  @$http
  @$translate
  @$q
  authentication
) ->
  # log out the user in case he entered via user settings
  authentication.clear()
  @

ForgotPasswordCtrl::submit = ->
  url = @endpoint + "/forgot-password"
  @$http.post(url, {@email}).then (response) =>
    @flash.success = @$translate.instant("forms.forgotPassword.flashes.success")
    @email = ""
    return
  .catch (response) =>
    @flash.error = @$translate.instant("forms.forgotPassword.flashes.fail")
    return

app.controller("forgotPasswordCtrl", ForgotPasswordCtrl)
