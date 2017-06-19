app = angular.module "harrowApp"

handleRequestError = ($state, error, origin) ->
  if error?.data?.reason and $state.get("errors/#{error.data.reason}")
    $state.go("errors/#{error.data.reason}", {error, origin})
  else if $state.get("errors/#{error?.status}")
    $state.go("errors/#{error?.status}", {error,origin})
  else
    console.error('Could not event redirect to error page', error)
    $state.go("errors/500", {error,origin})

resolves = ($transition$, token) ->
  if $transition$.getResolveTokens().includes(token)
    $transition$.getResolveValue(token)
  else
    return undefined

# Prevent $urlRouter from automatically intercepting URL changes;
# this allows you to configure custom behavior in between
# location changes and route synchronization:
app.config ($urlRouterProvider) ->
  $urlRouterProvider.otherwise("/a/dashboard")
  $urlRouterProvider.deferIntercept(true)

app.run ($q, feature, $log, $rootScope, $urlRouter, authentication, $state, $stateParams, $transitions, $trace, $parse, ic, ngProgressLite) ->
  # $trace.enable(1)
  $rootScope.$on '$locationChangeSuccess', (event) ->
    event.preventDefault()

    $q.all([feature.loadFeatures(), authentication.loadSession()]).then ->
      $urlRouter.sync()

  requiresAuthCriteria = {
    to: (state) ->
      state.data && state.data.requiresAuth
  }
  # NOTE: When route has `requiresAuth` then check session is valid
  $transitions.onStart requiresAuthCriteria, () ->
    if authentication.hasInvalidSession()
      $state.go('login', undefined, { location: false })

  # NOTE: When route has valid session and a parameter redirect to invitations
  $transitions.onStart {}, ($transition$) ->
    params = $transition$.params()
    if authentication.hasValidSession() && params.invitation
      $state.go('invitations/show', {uuid: params.invitation})

  $transitions.onStart {}, () ->
    ic.onTransition()

  $transitions.onStart {to: "login"}, () ->
    if authentication.hasValidSession()
      $state.go('dashboard', {returnTo: true})
    else if authentication.hasSession()
      $state.go('session_confirmation', $state.params)

  $transitions.onStart {to: "session_confirmation"}, () ->
    if authentication.hasValidSession()
      $state.go('dashboard', {returnTo: true})
    else if authentication.hasNoSession()
      $state.go('login', $state.params)

  $transitions.onStart {to: "signup"}, () ->
    if authentication.hasValidSession()
      $state.go('dashboard', {returnTo: true})
    else if authentication.hasSession() && authentication.hasInvalidSession()
      $state.go('session_confirmation', $state.params)

  $transitions.onSuccess {to: 'dashboard'}, ($transition$) ->
    organizations = resolves($transition$, 'organizations')
    projects = resolves($transition$, 'projects')
    if (organizations && projects) && (organizations.length == 0 && projects.length == 0)
      $state.go('wizard.quick-start')

  $transitions.onEnter {to: 'task.edit'}, ($transition$) ->
    params = $transition$.params()
    $state.go('task.edit.notification-rules', params)

  $transitions.onEnter {to: 'projects/edit'}, ($transition$) ->
    params = $transition$.params()
    $state.go('projects/edit.details', params)

  $transitions.onEnter {to: 'organization.edit'}, ($transition$) ->
    params = $transition$.params()
    $state.go('organization.edit.details', params)

  $transitions.onStart {}, ($transition$) ->
    params = $transition$.params('to')
    options = $transition$.options('to')

    if params.returnToHere == true
      newParams = angular.copy(params)
      newParams.returnTo = false
      newParams.returnToHere = false
      newParams.returnParams = angular.copy(params)
      newParams.returnState = $transition$.from().name
      $log.info('Transition: return point set "%s"', $transition$.from().name, newParams)
      return $state.go($transition$.to().name, newParams, options)

    if params.returnTo == true && params.returnState
      stateName = angular.copy(params.returnState)
      newParams = angular.copy(params.returnParams)
      newParams.returnTo = false
      newParams.returnToHere = false
      newParams.returnState = ''
      newParams.returnParams = {}
      $log.info('Transition: redirected back to return point "%s"', stateName, newParams)
      return $state.go(stateName, newParams, options)

  redirectToBilling = ($transition$) ->
    organization = resolves($transition$, 'organization')
    if $parse('_embedded.limits[0].subject.exceeded')(organization)
      return $state.go('billing', {uuid: organization.subject.uuid, organizationUuid: organization.subject.uuid})

  notBillingState = {
    to: (state) ->
      return state.name != 'billing' && state.name != 'createCreditCard' && state.name != 'organization.edit.archive' && state.name != 'projects/edit.archive'
  }

  $transitions.onEnter notBillingState, redirectToBilling

  $transitions.onSuccess notBillingState, redirectToBilling

  # ngProgress
  $transitions.onBefore {}, ->
    ngProgressLite.start()

  $transitions.onFinish {}, ->
    ngProgressLite.done()

  $transitions.onError {}, ->
    ngProgressLite.done()

  $rootScope.$on "$stateChangeSuccess", (event, toState, toParams, fromState, fromParams) ->
    ga?("set", "page", $state.href(toState.name, toParams))
    ga?("send", "pageview")

  # shows errors from resolvers, etc.
  $rootScope.$on "$stateNotFound", (event, toState, toParams, fromState, fromParams, error) ->
    $log.warn "$stateNotFound: %s", toState.to
    handleRequestError($state, error, location.hash)

  $rootScope.$on "$stateChangeError", (event, toState, toParams, fromState, fromParams, error) ->
    event.preventDefault()
    unless toState.name == "login"
      if error?.data?.reason == "session_expired" || !authentication.sessionUuid()
        authentication.clear()
        $state.go("login", {origin: location.hash})
        return
    handleRequestError($state, error, location.hash)
