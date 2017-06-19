Controller = (
  @project
  @triggers
  @tasks
  @environments
  @scripts
  @menuItems
  @$filter
  @task
  $controller
) ->
  $.extend(true, @, $controller('baseCtrl'))
  @



Controller::scriptFor = (trigger) ->
  task = @tasks.find (task) ->
    trigger.subject.taskUuid = task.subject.uuid
  if task
    script = @scripts.find (script) ->
      task.subject.scriptUuid = script.subject.uuid

Controller::environmentFor = (trigger) ->
  task = @itemFor(trigger.subject.taskUuid, 'tasks')
  @itemFor(task.subject.scriptUuid, 'environments')

Controller::editSrefFor = (triggerType) ->
  type = @$filter('singularize')(triggerType)
  if @task
    parent = "task.edit.triggers"
  else
    parent = "triggers"
  "#{parent}.#{type}.edit"

Controller::createSrefFor = (triggerType) ->
  type = @$filter('singularize')(triggerType)
  if @task
    parent = "task.edit.triggers"
  else
    parent = "triggers"
  "#{parent}.#{type}"

angular.module('harrowApp').controller 'triggersCtrl', Controller
