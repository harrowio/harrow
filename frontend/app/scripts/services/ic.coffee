IC = (
  @$http
  @$translate
  @authentication
  @flash
  @$q
  @intercomId
  $rootScope
) ->
  if @intercomId
    unwatch = $rootScope.$watch =>
      window.Intercom
    ,  =>
      unwatch()
      @boot()

  @queue=[]
  @

IC::boot = () ->
  @$http(
    method: 'OPTIONS'
    url: "https://api-ping.intercom.io/ping?app_id=#{@intercomId}"
    headers:
      "X-Harrow-Session-Uuid": undefined
  ).then =>
    Intercom('boot', @user())
    @available = true
    @afterBoot()
  .catch (res) =>
    # clear queue and make future enqueue's no-ops
    @queue = undefined
    @enqueue = (meth, params) =>
    @flash.error = @$translate.instant('intercom.blocked')

IC::enqueue = (meth, params) ->
  @queue.push {meth, params}

IC::afterBoot = () ->
  for call in @queue
    Intercom(call.meth,call.params)
  # replace enqueue with a non-queueing version
  @enqueue = (meth, params) =>
    Intercom(meth,params)

IC::newMessage = (msg) ->
  if !@available
    return false

  @enqueue "showNewMessage", msg
  return true

IC::onTransition = () ->
  @enqueue "update", @user()

IC::user = () ->
  app_id: @intercomId
  user_id: @authentication.currentUser?.subject?.uuid
  name: @authentication.currentUser?.subject?.name
  email: @authentication.currentUser?.subject?.email

angular.module('harrowApp').service 'ic', IC
