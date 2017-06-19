Controller = (
  @project
  @notifier
  @notifierType
  @notifierResource
  @tasks
  @scripts
  @environments
  @ga
  @flash
  @$translate
  @$q
  @$state
  @$stateParams
  @$injector
) ->
  @_generateNamesForTasks()
  @triggerActions = [
    'operation.succeeded'
    'operation.failed'
  ]
  if @notifierType == 'taskNotifier'
    @notifier.subject.triggeredByTaskUuid = @$state.params.taskUuid
  @nextState = 'notifiers'
  if @$state.current.data && @$state.current.data.nextState
    @nextState = @$state.current.data.nextState
  @

Controller::_generateNamesForTasks = () ->
  if @tasks
    @tasks.forEach (task) =>
      script = @scripts.find (script) ->
        script.subject.uuid == task.subject.scriptUuid
      env = @environments.find (env) ->
        env.subject.uuid == task.subject.environmentUuid
      if env && script
        task.subject.name = "#{env.subject.name} - #{script.subject.name}"

Controller::save = () ->
  @ga 'send', 'event', 'project#edit', 'notifiers', 'formSubmitted'
  @notifierResource.save(@notifier).then () =>
    @ga 'send', 'event', 'project#edit', 'notifiers', 'formSuccess'
    @flash.success = @$translate.instant("forms.notifiers.#{@notifierType}.flashes.create.success")
    @$state.go(@nextState, {projectUuid: @project.subject.uuid, returnTo: true}, {inherit: true, reload:true})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'project#edit', 'notifiers', 'formError'
    @flash.error = @$translate.instant("forms.notifiers.#{@notifierType}.flashes.create.fail")
    @$q.reject(reason)

Controller::delete = () ->
  @notifierResource.delete(@notifier.subject.uuid).then =>
    @ga 'send', 'event', 'project#edit', 'notifiers', 'deleteSuccess'
    @flash.success = @$translate.instant("forms.notifiers.#{@notifierType}.flashes.delete.success")
    @$state.go(@nextState, {projectUuid: @project.subject.uuid, returnTo: true}, {inherit: true, reload:true})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'project#edit', 'notifiers', 'deleteError'
    @flash.success = @$translate.instant("forms.notifiers.#{@notifierType}.flashes.delete.fail")


angular.module('harrowApp').controller 'notifierCtrl', Controller
