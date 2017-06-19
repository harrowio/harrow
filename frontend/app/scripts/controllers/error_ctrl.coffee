app = angular.module("harrowApp")

ErrorCtrl = (
  $stateParams
  @authentication
  @endpoint
  @flash
  @$state
  @$translate
  @$http
  @$q
) ->
  @error = $stateParams.error
  @origin = $stateParams.origin
  @

ErrorCtrl::resendVerificationEmail = () ->
  @$http(
    method: "POST"
    url: "#{@endpoint}/verify-email"
    headers:
      'X-Harrow-Session-Uuid': @authentication.currentSession.subject.uuid
  ).then () =>
    @$state.go("errors/verification_email_sent")
    return
  .catch (error) =>
    @flash.error = @$translate.instant("errors.blocked.flashes.failedToResendVerificationEmail")
    return

app.controller("errorCtrl", ErrorCtrl)
