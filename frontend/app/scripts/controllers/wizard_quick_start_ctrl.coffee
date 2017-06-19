Controller = (
  @$state
  @organizationResource
  @projectResource
  @repositoryResource
  @$filter
  Stateful
  @ga
) ->
  @ga 'send', 'event', 'quickStart', 'entered'
  @saveButtonOptions = {}
  @stateful = new Stateful()

  @stateful.on 'busy', =>
    @ga 'send', 'event', 'quickStart', 'saving'
    @saveButtonOptions =
      content: '<span svg-icon="icon-spinner"></span> Please Wait'
      attrs:
        class: 'btn'
        ngDisabled: true

  @stateful.on 'completed', =>
    @ga 'send', 'event', 'quickStart', 'completed'

  @stateful.on 'error', =>
    @ga 'send', 'event', 'quickStart', 'error'
    @saveButtonOptions =
      content: '<span svg-icon="icon-error-alt"></span> Error, Try Again?'
      attrs:
        class: "btn btn--primary"
        ngDisabled: false

  @ga 'send', 'event', 'quickStart', 'entered'

Controller::save = () ->
  @stateful.transitionTo('busy')
  url = @$filter('url')(@url, null)
  parts = url.pathname.replace(/^\//,'').split('/')
  isSsh = @$filter('isSsh')(@url)
  if parts.length == 2
    organizationName = @$filter('titlecase')(parts[0])
    projectName = @$filter('titlecase')(parts[1].replace(/\.git$/,''))
  else
    @stateful.transitionTo('error')
    @$state.go('wizard.create', {quickStartFailed: true})
    return

  @organizationResource.save(
    subject:
      public: false
      planUuid: "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
      name: organizationName
  )
  .then (organization) =>
    @organization = organization
    @projectResource.save(
      subject:
        organizationUuid: organization.subject.uuid
        name: projectName
    )
  .then (project) =>
    repository =
      subject:
        projectUuid: project.subject.uuid
    @project = project
    if isSsh
      repository.subject.url = @url.toString('ssh')
    else
      repository.subject.url = @url.toString('https')

    @repositoryResource.save(repository)
  .then (repository) =>
    @repository = repository
    repository.check().catch () =>
      return false
  .then (check) =>
    @stateful.transitionTo('completed')
    params = {projectUuid: @project.subject.uuid, repositoryUuid: @repository.subject.uuid}
    if check
      @$state.go('wizard.project.stencils', params)
    else if isSsh
      @$state.go('wizard.project.connect.repo.ssh', params)
    else
      @$state.go('wizard.project.connect.repo.private', params)
    return
  .catch (e) =>
    @stateful.transitionTo('error')


angular.module('harrowApp').controller 'wizardQuickStartCtrl', Controller
