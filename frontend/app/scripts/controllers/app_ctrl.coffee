app = angular.module("harrowApp")

AppCtrl = (
  @$rootScope
  @$window
  @$translate
  @authentication
  @flash
  @scheduleResource
  @operationResource
  @sessionResource
  @menuItems
  @$state
  @$q
  @ic
  @$filter
  @ws
  @modal
) ->
  @flash.subscribe (message, type) =>
    if type == 'success'
      @flash.error = null

  @$rootScope.$on "loggedIn", () =>
    @watchForSessionChanges(@authentication.sessionUuid())

  if sessionId = @authentication.sessionUuid()
    @watchForSessionChanges(sessionId)

  @$rootScope.$on 'http.forbidden', (event, statusCode, data) =>
    if data.reason == 'blocked'
      @$state.go('errors/blocked')
      return
    else if data.reason == 'session_invalidated'
      @$state.go('errors/session_invalidated')
      return
    else
      @flash.info = @$translate.instant("errors.http403.description")

  @

AppCtrl::watchForSessionChanges = (sessionId) ->
  cid = @ws.subRow "sessions", sessionId, =>
    @sessionResource.find(sessionId).then (session) =>
      if session.subject.invalidatedAt
        @ws.unsubscribe cid
        @$state.go("errors/session_invalidated")

AppCtrl::logout = ->
  @authentication.logout()
  @$state.go("login")

AppCtrl::back = ->
  @$window.history.back()

AppCtrl::isNew = () ->
  @$state.current.data.isNew == true

AppCtrl::zClipFail = ->
  alert("Sorry, we can't copy this to your clipboard, please check for blocked browser plugins.")

AppCtrl::runTaskNow = (task) ->
  @scheduleResource.save(
    subject:
      taskUuid: task.subject.uuid
      description: "Ad-hoc"
      timespec: "now"
  ).then =>
    @flash.success = @$translate.instant("tasks.flashes.runNow.success")
    @$state.go("task", {projectUuid: task.subject.projectUuid, taskUuid: task.subject.uuid}, {reload: true})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("tasks.flashes.runNow.fail")
    @$q.reject(reason)

AppCtrl::cancelOperation = (operation) ->
  @operationResource.delete(operation.subject.uuid).then =>
    @flash.success = @$translate.instant("operations.flashes.cancel.success")
  .catch (reason) =>
    @flash.error = @$translate.instant("operations.flashes.cancel.fail")
    @$q.reject(reason)

AppCtrl::openSupport = ($event, message = "Hi guys, I'm having problems with ...") ->
  if @ic.newMessage(message)
    $event.preventDefault()

AppCtrl::triggerState = (triggerType) ->
  type = @$filter('singularize')(triggerType)
  "task.edit.triggers.#{type}.edit"

AppCtrl::isAppSidebarStateActive = (state) ->
  if @$state.current.name == 'createRepository' && state == 'repositories'
    return true
  if @$state.current.name == 'createEnvironment' && state == 'environments'
    return true
  if @$state.current.name == 'invitations/create' && state == 'projects/edit.people'
    return true
  @$state.includes(state) || @$state.includes(@$filter('singularize')(state))

AppCtrl::taskEmailNotificationChange = (taskUuid) ->

AppCtrl::claimAccount = () ->
  @modal.show(
    templateUrl: 'views/modals/claim_account.html',
    mode: 'always',
    modal: 'claimAccount',
  ).then () ->
    window.location.reload()
  return

AppCtrl::isAuthenticated = () ->
  !!@authentication.currentUser

angular.module('harrowApp').controller("appCtrl", AppCtrl)
