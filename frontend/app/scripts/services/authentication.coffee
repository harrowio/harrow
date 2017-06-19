app = angular.module("harrowApp")

app.factory "authentication", (
  $http
  $q
  $filter
  userResource
  sessionResource
  uuid
  randomName
) ->
  Authentication = () ->
    @

  Authentication::localStorageSessionUuidKey = "Harrow-Session-Uuid"
  Authentication::localStorageSessionTempPasswordKey = "Harrow-Session-Temp-Password"
  Authentication::headerSessionUuidKey = "X-Harrow-Session-Uuid"

  Authentication::sessionUuid = () ->
    localStorage.getItem @localStorageSessionUuidKey

  Authentication::setSession = (session) ->
    localStorage.setItem(@localStorageSessionUuidKey, session.subject.uuid)
    @currentSession = session
    $http.defaults.headers.common[@headerSessionUuidKey] = session.subject.uuid
    @_getUser(session)
    .catch () =>
      @clear()
      $q.when()

  Authentication::_getUser = (session) ->
    session.user().then (user) =>
      if user && user.isGuest()
        user.subject.password = localStorage.getItem @localStorageSessionTempPasswordKey
      @currentUser = user

  Authentication::reloadCurrentUser = () ->
    sessionResource.find(@sessionUuid()).then (session) =>
      @setSession(session)

  Authentication::clear = () ->
    @currentUser = undefined
    @currentSession = undefined
    localStorage.removeItem(@localStorageSessionUuidKey)
    localStorage.removeItem(@localStorageSessionTempPasswordKey)
    delete $http.defaults.headers.common[@headerSessionUuidKey]

  # The returned promise always resolves, because for some controllers
  # (login, signup) having no currentSession is no problem
  Authentication::loadSession = () ->
    if @currentSession
      $q.when()
    else if @sessionUuid()
      sessionResource.find(@sessionUuid()).then (session) =>
        @setSession(session)
      .catch =>
        @clear()
        $q.when()
    else
      @clear()
      $q.when()

  Authentication::confirm = (totp) ->
    @currentSession.confirm(totp).then (session) =>
      @_getUser(session)
    .then (user) =>
      @currentUser = user

  Authentication::signup = (userProperties, signupParameters) ->
    userProperties.subject.signupParameters = signupParameters
    userResource.save(userProperties).then (newUser) =>
      @login
        subject:
          email: userProperties.subject.email
          password: userProperties.subject.password

  Authentication::userIsBlocked = () -> @currentSession?.subject.blocks?.length > 0

  Authentication::login = (sessionProperties) ->
    sessionResource.save(sessionProperties).then (session) =>
      localStorage.setItem(@localStorageSessionUuidKey, session.subject.uuid)
      @currentSession = session
      $http.defaults.headers.common[@headerSessionUuidKey] = session.subject.uuid
      @_getUser(session).then () =>
        session
      .catch ->
        session
    .catch (response) =>
      @clear() unless response?.data?.reason == "blocked"
      $q.reject(response)

  Authentication::guest = (signupParameters = {}) ->
    signupParameters.isGuest = true
    name = randomName()
    password = uuid()
    userProperties =
      subject:
        email: "guest-#{$filter('dashCase')(name)}-#{new Date().getTime()}@localhost"
        password: password
        name: name
    localStorage.setItem(@localStorageSessionTempPasswordKey, password)

    return @signup(userProperties, signupParameters)

  Authentication::isGuest = () ->
    return true if !@currentUser
    @currentUser.isGuest()

  Authentication::logout = () ->
    if @currentSession
      sessionResource.delete(@currentSession.subject.uuid)
      $q.when()
    @clear()

  # these methods need to return promises as they are used from resolve
  Authentication::hasNoSession = () ->
    !@currentSession || !!@currentSession.loggedOutAt

  Authentication::hasSession = () ->
    !@hasNoSession()

  Authentication::hasValidSession = () ->
    @hasSession() && @currentSession.subject.valid

  Authentication::hasInvalidSession = () ->
    !@hasValidSession()

  new Authentication()
