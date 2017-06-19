app = angular.module("harrowApp")

TaskListShowCtrl = (
  @project
  @tasks
  @environments
  @scripts
  @taskResource
  @$translate
  @$state
  @$scope
  @flash
  @$q
) ->
  @currentState = @$state.current
  @

TaskListShowCtrl::environmentFor = (task) ->
  environments = @environments.filter (env) ->
    env.subject.uuid == task.subject.environmentUuid
  environments[0] if environments.length

TaskListShowCtrl::scriptFor = (task) ->
  scripts = @scripts.filter (script) ->
    script.subject.uuid == task.subject.scriptUuid
  scripts[0] if scripts.length

app.controller("taskListShowCtrl", TaskListShowCtrl)
