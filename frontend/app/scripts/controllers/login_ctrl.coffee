app = angular.module("harrowApp")

LoginCtrl = (
  @$q
  @$state
  @$stateParams
  @$rootScope
  @$translate
  @flash
  @authentication
  @ga
  @oauth
  Stateful
) ->
  @stateful = new Stateful()
  @stateful.on 'busy', =>
    @buttonStateful =
      content: '<span svg-icon="icon-spinner" class="iconColor"></span> Logging in'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'success', =>
    @buttonStateful =
      content: '<span svg-icon="icon-spinner" class="iconColor"></span> Logged in, please wait'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'error', =>
    @buttonStateful =
      content: 'Error, Try again?'
      attrs:
        class: 'btn btn--primary'
        ngDisabled: false
  @ga 'send', 'event', 'user', 'loginForm'
  @invitationUuid = @$stateParams.invitation
  @origin = @$stateParams.origin
  @

LoginCtrl::login = () ->
  @stateful.transitionTo('busy')
  @authentication.login(@user).then () =>
    @stateful.transitionTo('success')
    @ga 'send', 'event', 'user', 'login-ok'

    if @authentication.userIsBlocked()
      @$state.go("errors/blocked")
      return
    unless @authentication.hasValidSession()
      @$state.go("session_confirmation", @$stateParams)
    else
      @$rootScope.$emit "loggedIn"
      if @invitationUuid
        @$state.go("invitations/show", {uuid: @invitationUuid})
      else
        if @origin
          window.location.hash = @origin
        else
          @$state.go("dashboard")
    return
  .catch (reason) =>
    @stateful.transitionTo('error')
    @ga 'send', 'event', 'user', 'login-failed'
    @flash.error = @$translate.instant("forms.login.flashes.loginFailed")
    @$q.reject(reason)

LoginCtrl::guest = () ->
  @stateful.transitionTo('busy')
  @authentication.guest(@$stateParams).then () =>
    @stateful.transitionTo('success')
    @ga 'send', 'event', 'user', 'guest-ok'
    @$state.go('dashboard')
    return
  .catch (reason) =>
    @stateful.transitionTo('error')
    @ga 'send', 'event', 'user', 'guest-failed'
    @flash.error = @$translate.instant('forms.login.flashes.guestFailed')
    @$q.reject reason




LoginCtrl::githubSignin = () ->
  @oauth.signinGithub()

app.controller("loginCtrl", LoginCtrl)
