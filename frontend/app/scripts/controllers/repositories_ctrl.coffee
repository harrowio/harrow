Controller = (
  @repositories
  @project
  @$filter
  @Stateful
) ->
  @state = {}
  @statefulOptions = {}
  @repositories.forEach (repository) =>
    @_generateStateful(repository)
    if repository.subject.accessible
      @_setState(repository.subject.uuid, 'accessible')
    else
      @_setState(repository.subject.uuid, 'inaccessible')
  @

Controller::_generateStateful = (repository) ->
  stateful = new @Stateful()
  options = {}
  options.attrs =
    svgIcon: 'icon-repositories'

  stateful.on 'busy', ->
    options.attrs.svgIcon = 'icon-spinner'
    options.attrs.class = 'iconColor'

  stateful.on 'accessible', ->
    options.attrs.svgIcon = 'icon-complete'
    options.attrs.class = 'iconColor'

  stateful.on 'inaccessible', ->
    options.attrs.svgIcon = 'icon-error'
    options.attrs.class = 'iconColor'

  @statefulOptions[repository.subject.uuid] = options
  @state[repository.subject.uuid] = stateful

Controller::_setState = (uuid, state) ->
  opts = {}
  if state != 'busy'
    opts.terminal = true
  @state[uuid].transitionTo(state, opts)
  return

Controller::checkNow = (repository) ->
  @_setState(repository.subject.uuid, 'busy')
  repository.check('update-metadata').then () =>
    @_setState(repository.subject.uuid, 'accessible')
    return
  .catch () =>
    @_setState(repository.subject.uuid, 'inaccessible')
    return


Controller::isPrivate = (repository) ->
  @$filter('url')(repository.subject.url).isPrivate

angular.module('harrowApp').controller 'repositoriesCtrl', Controller
