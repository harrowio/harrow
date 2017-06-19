Controller = (
  @project
  @notifiers
  @scripts
  @environments
  @tasks
  @$filter
  @menuItems
) ->
  @_generateTaskNames()
  @


Controller::_taskName = (task) ->
  script = @scripts.find (script) ->
    script.subject.uuid == task.subject.scriptUuid
  env = @environments.find (env) ->
    env.subject.uuid == task.subject.environmentUuid
  if script and env
    "#{env.subject.name} - #{script.subject.name}"

Controller::_generateTaskNames = () ->
  if angular.isArray(@notifiers.taskNotifiers)
    @notifiers.taskNotifiers.forEach (item) =>
      task = @tasks.find (task) ->
        task.subject.uuid == item.subject.taskUuid
      triggerTask = @tasks.find (task) ->
        task.subject.uuid == item.subject.triggeredByTaskUuid
      item.subject.taskName = @_taskName(task)
      item.subject.triggerTaskName = @_taskName(triggerTask)

Controller::editSrefFor = (type) ->
  type = @$filter('singularize')(type)
  "notifiers.#{type}.edit"

Controller::createSrefFor = (type) ->
  type = @$filter('singularize')(type)
  "notifiers.#{type}"

angular.module('harrowApp').controller 'notifiersCtrl', Controller
