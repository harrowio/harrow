Controller = (
  @ga
  @organizationResource
  @projectResource
  @organization
  @project
  @flash
  @$translate
  @$state
  @$q
) ->
  @ga 'send', 'event', 'wizard', 'create', 'entered'
  @

Controller::save = () ->
  @ga 'send', 'event', 'wizard', 'create', 'formSubmitted'
  promise = if @project.subject.organizationUuid
    @_saveProject()
  else
    @_saveOrg()
  promise.then (project) =>
    @ga 'send', 'event', 'wizard', 'create', 'formSuccess'
    @flash.success = @$translate.instant("forms.wizard.create.flashes.success", project.subject)
    @$state.go('wizard.project.connect', {projectUuid: project.subject.uuid}, {reload: true})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'wizard', 'create', 'formError'
    @flash.error = @$translate.instant("forms.wizard.create.flashes.fail", @project.subject)

    @$q.reject(reason)

Controller::_saveOrg = () ->
  @organizationResource.save(@organization).then (organization) =>
    @project.subject.organizationUuid = organization.subject.uuid
    @_saveProject()

Controller::_saveProject = () ->
  @projectResource.save(@project).then (project) =>
    @project = project
    project

angular.module('harrowApp').controller 'wizardCreateCtrl', Controller
