app = angular.module("harrowApp")

VerifyEmailCtrl = (
  @endpoint
  @userUuid
  @token
  @flash
  @$http
  @$state
  @$translate
  @$q
) ->
  @verify()
  @

VerifyEmailCtrl::verify = ->
  url = @endpoint + "/users/#{@userUuid}/verify-email"
  req = {@token}
  @$http.post(url, req).success (response) =>
    @flash.success = @$translate.instant("forms.verifyEmail.flashes.success")
    @$state.go("dashboard")
    return
  .error (response) =>
    @flash.error = @$translate.instant("forms.verifyEmail.flashes.fail")
    return

app.controller("verifyEmailCtrl", VerifyEmailCtrl)
