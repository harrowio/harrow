app = angular.module("harrowApp")

SignupCtrl = (
  @authentication
  @flash
  @$q
  @$translate
  @$state
  @$stateParams
  @ga
  @oauth
  Stateful
) ->
  @stateful = new Stateful()
  @stateful.on 'busy', =>
    @buttonStateful =
      content: '<span svg-icon="icon-spinner" class="iconColor"></span> Creating Account'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'success', =>
    @buttonStateful =
      content: '<span svg-icon="icon-spinner" class="iconColor"></span> Created Account, please wait'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'error', =>
    @buttonStateful =
      content: 'Error, Try again?'
      attrs:
        class: 'btn btn--primary'
        ngDisabled: false
  @ga 'send', 'event', 'user', 'signupForm'

  @invitationUuid = @$stateParams.invitation
  @origin = @$stateParams.origin
  @

SignupCtrl::signup = () ->
  @stateful.transitionTo('busy')

  if @invitationUuid
    @user.subject.invitationUuid = @invitationUuid

  @authentication.signup(@user, @$stateParams).then () =>
    @stateful.transitionTo('success')
    @flash.success = @$translate.instant("forms.signup.flashes.success")
    @ga 'send', 'event', 'user', 'signup-ok'
    if @authentication.userIsBlocked()
      @$state.go("errors/blocked")
      return
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
    @ga 'send', 'event', 'user', 'signup-failed'
    @flash.error = @$translate.instant("forms.signup.flashes.fail")
    @$q.reject(reason)

SignupCtrl::guest = () ->
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

SignupCtrl::githubSignin = () ->
  @oauth.signinGithub()

app.controller("signupCtrl", SignupCtrl)
