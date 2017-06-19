app = angular.module("harrowApp")

GitTriggerListCtrl = (
  @$controller
  @tasks
  @environments
  @scripts
  @repositories
  @project
  @gitTriggers
  @gitTriggerResource
  @flash
  @$translate
) ->
  $.extend(true, @, $controller('baseCtrl'))

  @taskNames = {}

  for task in @tasks
    @taskNames[task.subject.uuid] = task.subject.name

  @repositoryNames =
    "": "any repository"

  for repository in @repositories
    @repositoryNames[repository.subject.uuid] = repository.subject.name

  @

GitTriggerListCtrl::delete = (gitTrigger) ->
  if confirm(@$translate.instant("prompts.really?"))
    @gitTriggerResource.delete(gitTrigger.subject.uuid).then =>
      @flash.success = @$translate.instant("gitTriggers.flashes.delete.success", gitTrigger.subject)
      @gitTriggers = @gitTriggers.filter( (hook) -> hook.subject.uuid != gitTrigger.subject.uuid )
      return
    .catch =>
      @flash.error = @$translate.instant("gitTriggers.flashes.delete.failure", gitTrigger.subject)
      return

GitTriggerListCtrl::environmentFor = (hook) ->
  @itemFor(@itemFor(hook.subject.taskUuid, 'tasks').subject.environmentUuid, 'environments')

GitTriggerListCtrl::scriptFor = (hook) ->
  @itemFor(@itemFor(hook.subject.taskUuid, 'tasks').subject.scriptUuid, 'scripts')


app.controller("gitTriggerListCtrl", GitTriggerListCtrl)
