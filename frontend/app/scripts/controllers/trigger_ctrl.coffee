Controller = (
  @project
  @triggerType
  @trigger
  @triggerResource
  @flash
  @repositories
  @tasks
  @scripts
  @environments
  @task
  @ga
  @$translate
  @$q
  @$state
  @$stateParams
) ->
  @_generateNamesForTasks()
  @scheduleType = if @trigger.subject.cronspec then 'cronspec' else 'timespec'
  @repositoryOptions = []
  @repositoryOptions.push { url: 'all', uuid: null }
  @hasRepositoryIssue = false
  @repositories.forEach (repository) =>
    unless repository.subject.accessible
      @hasRepositoryIssue = true
    @repositoryOptions.push {
      url: repository.subject.url
      uuid: repository.subject.uuid
    }
  @nextState = 'triggers'
  if @$state.current.data && @$state.current.data.nextState
    @nextState = @$state.current.data.nextState

  if @$state.current.data?.autoSave
    @save()
  @

Controller::_generateNamesForTasks = () ->
  @tasks.forEach (task) =>
    script = @scripts.find (script) ->
      script.subject.uuid == task.subject.scriptUuid
    env = @environments.find (env) ->
      env.subject.uuid == task.subject.environmentUuid
    if env && script
      task.subject.name = "#{env.subject.name} - #{script.subject.name}"

Controller::save = () ->
  @ga 'send', 'event', 'triggers', @triggerType, 'formSubmitted'
  @triggerResource.save(@trigger).then () =>
    @ga 'send', 'event', 'triggers', @triggerType, 'formSuccess'
    @flash.success = @$translate.instant("forms.triggers.#{@triggerType}.flashes.success")
    @$state.go(@nextState, { projectUuid: @project.subject.uuid, returnTo: true }, {reload: true})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'triggers', @triggerType, 'formError'
    @flash.error = @$translate.instant("forms.triggers.#{@triggerType}.flashes.fail")
    @$q.reject(reason)

Controller::delete = () ->
  @triggerResource.delete(@trigger.subject.uuid).then =>
    @ga 'send', 'event', 'project#edit', 'triggers', 'deleteSuccess'
    @flash.success = @$translate.instant("forms.triggers.#{@triggerType}.flashes.delete.success")
    @$state.go(@nextState, {projectUuid: @project.subject.uuid, returnTo: true}, {inherit: true, reload:true})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'project#edit', 'triggers', 'deleteError'
    @flash.success = @$translate.instant("forms.triggers.#{@triggerType}.flashes.delete.fail")

angular.module('harrowApp').controller 'triggerCtrl', Controller
