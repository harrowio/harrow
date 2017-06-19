Controller = (
  @operations
  @environments
  @repositories
  @tasks
  @scripts
) ->
  @_generateNames()
  @

Controller::_generateNames = () ->
  @operations.forEach (item) =>
    task = @tasks.find (task) ->
      item.subject.taskUuid == task.subject.uuid
    if task
      script = @scripts.find (script) ->
        task.subject.scriptUuid == script.subject.uuid
      env = @environments.find (env) ->
        task.subject.environmentUuid == env.subject.uuid
      if script && env
        item.subject.environmentName = env.subject.name
        item.subject.scriptName = script.subject.name
        item.subject.name = "#{env.subject.name} - #{script.subject.name}"

Controller::repositoryFor = (obj) ->
  uuid = if obj.subject then obj.subject.repositoryUuid else obj
  items = @repositories.filter (item) ->
    item.subject.uuid == uuid
  items[0] if items.length

angular.module('harrowApp').controller 'operationListCtrl', Controller
