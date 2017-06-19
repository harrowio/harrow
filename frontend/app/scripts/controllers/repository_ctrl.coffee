Controller = (
  @project
  @repository
  @repositoryResource
  @credential
  @ga
  @flash
  @$timeout
  @$translate
  @$state
  @$filter
  @$scope
  @ws
  @$q
  @$log
  Stateful
) ->
  @stateful = new Stateful()
  @checkPromise = null
  @sockets = []
  $scope.$on '$destroy', =>
    @sockets.forEach (socket) =>
      @ws.unsubscribe socket

  if @repository.subject.uuid
    @sockets.push @ws.subRow "repositories", @repository.subject.uuid, () =>
      @repositoryResource.find(@repository.subject.uuid).then (repository) =>
        @repository = repository
        @$timeout =>
          @$scope.$apply()

  if @credential
    @sockets.push @ws.subRow 'repository_credentials', @credential.subject.uuid, () =>
      @$log.debug 'Have a response from repository_credentials via WebSocket'
      @repository.credential().then (credential) =>
        @credential = credential
      @$timeout =>
        @$scope.$apply()

  @stateful.on 'checking', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-spinner"></span> Checking'
      attrs:
        class: 'btn'
        ngDisabled: true

  @stateful.on 'accessible', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-complete-alt"></span> Accessible'
      attrs:
        class: 'btn btn--green'

  @stateful.on 'private', =>
    @saveButtonOptions =
      content: 'More Detail Required'
      attrs:
        class: 'btn btn--yellow'
        ngDisabled: false

  @stateful.on 'complete', =>
    @saveButtonOptions =
      content: 'Fantastic!'
      attrs:
        class: 'btn btn--green'

  @stateful.on 'error', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-error-alt"></span> Try again?'
      attrs:
        class: "btn btn--primary"
        ngDisabled: false

  @stateful.on 'softTimeout', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-spinner"></span> Not too long now'
      attrs:
        class: 'btn'
        ngDisabled: true

  @stateful.on 'timeout', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-error-alt"></span> Timeout, Try again?'
      attrs:
        class: "btn btn--blue"
        ngDisabled: false

  @url = @$filter('url')(@repository.subject.url)
  @finalState = @$state.current.data.nextState.accessible
  @

Controller::isHttps = () ->
  !@isSsh()

Controller::isSsh = () ->
  @$filter('isSsh')(@repository.subject.url)

Controller::nameFromUrl = () ->
  @url = @$filter('url')(@repository.subject.url)
  return @url.pathname.replace('.git', '')

Controller::_nextState = (accessible) ->
  if accessible
    @flash.success = @$translate.instant(
      'forms.repository.flashes.save.success',
      @repository.subject
    )
    state = @finalState
  else
    if @isSsh()
      state = @$state.current.data.nextState.ssh
    else
      state = @$state.current.data.nextState.https

  @$state.go(state, {
    projectUuid: @project.subject.uuid
    repositoryUuid: @repository.subject.uuid
  },
  {reload: true})

Controller::_check = () ->
  @repository.check()

Controller::save = ->
  @stateful.transitionTo('checking')

  @repository.subject.name = @nameFromUrl()
  if @isSsh()
    @_saveSsh()
  else
    @_saveHttps()

Controller::_saveHttps = ->
  @repositoryResource.save(@repository).then (repository) =>
    @repository = repository
    return repository
  .then (repository) =>
    return repository.check().catch () ->
      return false
  .then (accessible) =>
    if accessible
      @stateful.transitionTo('complete', {terminal: true})
      @_nextState(true)
    else
      @stateful.transitionTo('private', {terminal: true})
      @_nextState(false)
    return
  .catch =>
    @stateful.transitionTo('error', {terminal: true})
    return

Controller::_saveSsh = ->
  @repositoryResource.save(@repository)
  .then (repository) =>
    @repository = repository
    @stateful.transitionTo('private')
    @_nextState(false)
    return

Controller::check = () ->
  @stateful.transitionTo('checking')
  if @isSsh(@repository.subject.url)
    @repository.subject.url = @url.toString('ssh')
  else
    @repository.subject.url = @url.toString('https')
  @repositoryResource.save(@repository).then () =>
    return @repository.check()
  .then (check) =>
    @stateful.transitionTo('complete', {terminal: true})
    @_nextState(true)
    return
  .catch () =>
    @stateful.transitionTo('error', {terminal: true})
    return

Controller::delete = () ->
  @repositoryResource.delete(@repository.subject.uuid).then () =>
    @flash.success = @$translate.instant(
      'forms.repository.flashes.delete.success',
      @repository.subject
    )

    @_nextState(true)
    return

angular.module('harrowApp').controller 'repositoryCtrl', Controller
