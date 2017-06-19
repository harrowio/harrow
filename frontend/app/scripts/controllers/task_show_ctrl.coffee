app = angular.module("harrowApp")

TaskShowCtrl = (
  @task
  @script
  @environment
  @triggers
  @notifiers
  @project
  @$state
  @$translate
  @flash
  @authentication
  @operations
  $window
  @$scope
  initEvents
) ->
  # TODO: hack, we should provide a way to subscribe to table modification (i.e. new record) in addition to row
  # subscriptions
  il = $window.setInterval =>
    @refreshOperations()
  , 5000
  @$scope.$on "$destroy", ->
    $window.clearInterval il

  @filterTriggerableTasks()
  @triggerTaskAction = 'succeeded'
  @watching = false
  angular.forEach(@notificationRules, (rule) =>
    if rule.subject && rule.subject.creatorUuid == authentication?.currentUser?.subject.uuid
      @watching = true
  )
  initEvents(@, @$scope)

  @

TaskShowCtrl::events = [
  "taskControlsRun",
  "taskNotifierAdded",
  "taskNotifierRemoved",
]

TaskShowCtrl::hasTriggers = () ->
  arr = []
  Object.keys(@triggers).forEach (trigger) =>
    arr.push @triggers[trigger].length
  arr.some (item) ->
    item > 0

TaskShowCtrl::hasNotifiers = () ->
  arr = []
  Object.keys(@notifiers).forEach (item) =>
    arr.push @notifiers[item].length
  arr.some (item) ->
    item > 0

TaskShowCtrl::taskControlsRunHandler = (_, p) ->
  p.then () =>
    @refreshOperations()


TaskShowCtrl::filterTriggerableTasks = () ->
  @triggerableTasks = []
  angular.forEach(@tasks, (task) =>
    if task.subject.uuid == @task.subject.uuid
      return

    @triggerableTasks.push(task)
  )
TaskShowCtrl::createSchedule = () ->
  @scheduleResource.save
    subject:
      taskUuid: @task.subject.uuid
      description: "Re-run of operation " + @operation.subject.uuid
      timespec: "now"

TaskShowCtrl::watch = () ->
  @task.watch().then () =>
    @watching = true
    return
  .catch () =>
    @watching = false
    return

TaskShowCtrl::unwatch = () ->
  @watching = false
  @task.unwatch()

TaskShowCtrl::delete = () ->
  if @confirm()
    @taskResource.delete(@task.subject.uuid).then =>
      @$state.go("projects/edit", {uuid: @project.subject.uuid})
      return

TaskShowCtrl::refreshOperations = () ->
  @task.operations().then (operations) =>
    @operations = operations
    return

TaskShowCtrl::confirm = () ->
  confirm(@$translate.instant("prompts.really?"))

TaskShowCtrl::deleteTaskNotifier = (taskNotifier) ->
  if @confirm()
    @taskNotifierResource.delete(taskNotifier.subject.uuid).then =>
      @$scope.$emit("taskNotifierAdded")
      return

TaskShowCtrl::createTaskNotifier = () ->
  @taskNotifierResource.save
    subject:
      taskUuid: @triggerTask
      triggeredByTaskUuid: @task.subject.uuid
      triggerAction: @triggerTaskAction
  .then () =>
    @$scope.$emit("taskNotifierAdded")

TaskShowCtrl::taskNotifierTriggerAction = (taskNotifier) ->
  switch taskNotifier.subject.triggerAction
    when 'operation.succeeded' then 'success'
    when 'operation.failed' then 'failure'
    else '?'


TaskShowCtrl::mostRecentOperationFor = (task) ->
  if task
    operations = @operations.filter (operation) ->
      operation.subject.jobUuid == task.subject.uuid
    operations = operations.sort (a, b) ->
      if a.subject.createdAt > b.subject.createdAt
        return -1
      else if  a.subject.createdAt < b.subject.createdAt
        return 1
      else
        return 0
    operations[0]

app.controller("taskShowCtrl", TaskShowCtrl)
