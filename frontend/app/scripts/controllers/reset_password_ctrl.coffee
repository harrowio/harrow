app = angular.module("harrowApp")

ResetPasswordCtrl = (
  @endpoint
  @mac
  @email
  @authentication
  @flash
  @$http
  @$state
  @$translate
) ->
  @

ResetPasswordCtrl::submit = ->
  url = @endpoint + "/reset-password"
  req = {@email, @mac, @password}
  @$http.post(url, req).then (response) =>
    @flash.success = @$translate.instant("forms.resetPassword.flashes.success")
    subject:
      email: req.email
      password: req.password
  .then (loginReq) =>
    @authentication.login(loginReq)
  .then (session) =>
    if session.subject.valid
      @$state.go("dashboard")
    else
      @$state.go("session_confirmation")
    return
  .catch (response) =>
    @flash.error = @$translate.instant("forms.resetPassword.flashes.fail")
    return

app.controller("resetPasswordCtrl", ResetPasswordCtrl)
