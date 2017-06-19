app = angular.module("harrowApp")

Controller = (
  @$scope
  @$element
  @ws
  @$window
  @$log
) ->
  @sockets = []
  @followLogs = @$scope.followLogs || false
  @streaming = false

  @target = @$element.find("pre")[0]

  @$scope.$on "$destroy", () =>
    @_unsubscribeFromEventLogs()

  @limitedHeightScroll = angular.element('.lloogg').css('maxHeight') != 'none'
  @_initWatchers()
  @

Controller::_unsubscribeFromEventLogs = ->
  if @sockets.length > 0
    @sockets.forEach (socket) =>
      @$log.debug('LogView: Unsubscribed LogEvent', socket)
      @ws.unsubscribe(socket)

Controller::_subscribeToLogEvents = (uuid) ->
  @_unsubscribeFromEventLogs()
  @lom = new Lom.Lom(@$window.document, Lom.defaultHandlers, 1000)
  socket = @ws.subLogevents uuid, (update) =>
    @handleLogEvents(update)
  @$log.debug('LogView: Subscribed LogEvent (%s)', uuid, socket)
  @sockets.push socket

Controller::_initWatchers = () ->
  @$scope.$watch () =>
    @followLogs
  , (newVal, oldVal) =>
    @follow() if newVal

  @$scope.$watch 'operationUuid', (uuid) =>
    @$log.debug 'LogView: uuidChanged', uuid
    @target.innerHTML = 'Please wait, logs will appear here soon...'
    @streaming = false
    if uuid
      @_subscribeToLogEvents(uuid)

Controller::handleLogEvents = (loggable) ->
  if @streaming == false
    @target.innerHTML = ''
    @streaming = true
  for ev in loggable.logevents
    event_data = ev
    event_data = ev.E if ev.E
    @lom.pushEvent event_data

  if !@hasEverRendered
    @target.innerHTML = ""
    @hasEverRendered = true
  @lom.render @target
  @follow() if @followLogs
  last_event = loggable.logevents[loggable.logevents.length - 1]
  if (last_event.type || last_event.E.type) == "eof"
    @_unsubscribeFromEventLogs()

Controller::follow = () ->
  if @limitedHeightScroll == true
    container = angular.element('.lloogg')
  else
    container = angular.element('.app__container')

  height = Math.ceil(container.prop('scrollHeight'))
  container.scrollTop(height)

app.controller 'logViewCtrl', Controller

app.directive 'logView', ($rootScope, $window, ws) ->
  logBody = """
  <div class="llooggContainer">
    <div class="lloogg-toggleFollow lloogg-toggleFollow-top">
      <div class="field__checkbox">
        <label>
          <input type="checkbox" ng-click="ctrl.followLogs = !ctrl.followLogs" ng-checked="ctrl.followLogs">
          <span></span>
          Scroll with log
        </label>
      </div>
    </div>
    <pre class="lloogg">Please wait, logs will appear here soon...</pre>
    <div class="lloogg-toggleFollow lloogg-toggleFollow-bottom">
      <div class="field__checkbox">
        <label>
          <input type="checkbox" ng-click="ctrl.followLogs = !ctrl.followLogs" ng-checked="ctrl.followLogs">
          <span></span>
          Scroll with log
        </label>
      </div>
    </div>
  </div>
  """
  {
    restrict: 'E'
    template: logBody
    controller: Controller
    controllerAs: 'ctrl'
    scope:
      operationUuid: '@operationUuid'
      followLogs: '@followLogs'
    replace: true
  }
