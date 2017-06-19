Controller = (
  @script
  @testScript
  @project
  @environments
  @repositories
  @secrets
  @scriptResource
  @scriptEditorResource
  @operationResource
  @taskResource
  @flash
  @$translate
  @$state
  @ws
  @$scope
  @$q
  @$log
  Stateful
) ->
  @stateful = new Stateful()
  @isDirty = false
  @isPristine = true
  @operation = null
  @sockets = []
  @tabView = 'environment'

  @buttonStates = {}

  @stateful.on 'idle', =>
    @buttonStates.testScript =
      content: 'Test Script'
      attrs:
        class: 'btn btn--primary'
        ngClick: 'ctrl.test()'
        ngDisabled: 'harrowForm.form.$invalid'

  @stateful.on 'starting', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-spinner"></span> Starting - Cancel'
      attrs:
        class: 'btn'
        ngClick: 'appctrl.cancelOperation(ctrl.operation)'
        ngDisabled: false

  @stateful.on 'softTimeout', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-spinner"></span> Waiting for VM - Cancel'
      attrs:
        class: 'btn'
        ngClick: 'appctrl.cancelOperation(ctrl.operation)'
        ngDisabled: false

  @stateful.on 'running', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-spinner"></span> Running - Cancel'
      attrs:
        class: 'btn'
        ngClick: 'appctrl.cancelOperation(ctrl.operation)'
        ngDisabled: false

  @stateful.on 'success', =>
    if @isPristine
      @buttonStates.testScript =
        content: 'Success, Save Script'
        attrs:
          class: 'btn btn--green'
          ngDisabled: false
          ngClick: 'ctrl.save()'
    else
      @buttonStates.testScript =
        content: 'Test Script, Again'
        attrs:
          class: 'btn btn--primary'
          ngClick: 'ctrl.test()'
          ngDisabled: 'harrowForm.form.$invalid'

  @stateful.on 'timeout', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-error"></span> Timeout'
      attrs:
        class: 'btn btn--blue'
        ngDisabled: false
        ngClick: 'appctrl.openSupport("Help! I just observed a timeout while testing a script")'

  @stateful.on 'fatal', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-error"></span> Fatal Error'
      attrs:
        class: 'btn btn--black'
        ngDisabled: false
        ngClick: 'appctrl.openSupport("Help! I just observed a fatal error while testing a script")'

  @stateful.on 'canceled', =>
    @buttonStates.testScript =
      content: 'Cancelled, Try Again?'
      attrs:
        class: 'btn btn--red'
        ngClick: 'ctrl.test()'
        ngDisabled: false

  @stateful.on 'error', =>
    @buttonStates.testScript =
      content: '<span svg-icon="icon-error"></span> Failed, Try Again?'
      attrs:
        class: 'btn btn--red'
        ngDisabled: false

  @variables = []
  @secretVariables = []
  @sshKeys = []

  @selectedEnvironment = angular.copy(@testScript.environment.uuid)
  @chooseEnvironment()

  @$scope.$watch () =>
    @testScript.script.body
  , (current, previous) =>
    @_setDirty()

  @$scope.$watch () =>
    @variables
  , (current, previous) =>
    @_setDirty()
    @variables = @_addEmptyVariable(current)
  , true

  @$scope.$watch () =>
    @secretVariables
  , (current, previous) =>
    @_setDirty()
    @secretVariables = @_addEmptyVariable(current)
  , true

  @$scope.$watch () =>
    @sshKeys
  , (current, previous) =>
    if current != previous
      @_setDirty()
    @sshKeys = @_addEmptyVariable(current)
  , true

  @$scope.$on 'uiAce:textChanged', (event, value) =>
    @$scope.$apply () =>
      @testScript.script.body = value

  @$scope.$on '$destroy', =>
    @sockets.forEach (socket) =>
      @ws.unsubscribe socket
  @

Controller::chooseEnvironment = () ->
  environment = @environments.find (env) =>
    env.subject.uuid == @selectedEnvironment
  .subject
  @testScript.environment = environment
  @_setDirty()
  @variables = []
  Object.keys(environment.variables).forEach (key) =>
    @variables.push name: key, value: environment.variables[key]
  @secretVariables = []
  @sshKeys = []
  @secrets.filter (secret) ->
    secret.subject.environmentUuid == environment.uuid
  .forEach (secret) =>
    if secret.subject.type == 'env'
      @secretVariables.push name: secret.subject.name, value: secret.subject.value
    if secret.subject.type == 'ssh'
      @sshKeys.push name: secret.subject.name, value: secret.subject.value

Controller::aceLoaded = (_editor) ->
  # See: https://github.com/angular-ui/ui-ace/issues/104
  _editor.$blockScrolling = Infinity
  setTimeout ->
    _editor.resize()
  ,0

Controller::_addEmptyVariable = (variables) ->
  populated = variables.filter (variable) ->
    angular.isDefined(variable.name) && variable.name != ''
  populated.push({})
  populated

Controller::_convertArrayToKeyPair = (newVariables) ->
  keyPair = {}
  newVariables.forEach (variable) ->
    if variable.name
      keyPair[variable.name] = variable.value

  keyPair

Controller::_setDirty = () ->
  @isDirty = true
  @isPristine = false
  if @stateful.isTerminal && @stateful.state != 'running'
    @stateful.transitionTo('idle', {terminal: true})

Controller::_setPristine = () ->
  @isDirty = false
  @isPristine = true

Controller::_subscribeToOperation = (operation) ->
  @$log.debug('ScriptEditorCtrl::_subscribeToOperation: %s', operation.subject.uuid)
  uuid = operation.subject.uuid

  @sockets.push @ws.subRow 'operations', uuid, (event, data) =>
    @$log.debug('ScriptEditorCtrl::_subscribeToOperation~~eventData', data)
    @operationResource.find(uuid).then (response) =>
      @$log.debug("ScriptEditorCtrl::OperationResource: %s - Status: ", uuid, response.status())
      @operation = response
      if response.status() == 'success'
        @stateful.transitionTo('success', {terminal: true})
      else if response.status() == 'running' || response.status() == 'pending'
        @stateful.transitionTo('running', {terminal: true})
      else if response.status() == 'fatalerr'
        @stateful.transitionTo('fatal', {terminal: true})
      else if response.status() == 'timedout'
        @stateful.transitionTo('timeout', {terminal: true})
      else if response.status() == 'canceled'
        @stateful.transitionTo('canceled', {terminal: true})
      else
        @stateful.transitionTo('error', {terminal: true})

Controller::_updateTestScript = () ->
  secrets = []
  @secretVariables.forEach (item) ->
    if item.name
      secrets.push {name: item.name, value: item.value}
  @testScript.secrets = secrets
  @testScript.environment.variables = @_convertArrayToKeyPair(@variables)

Controller::test = () ->
  @tabView = 'console'
  @_updateTestScript()
  @stateful.transitionTo('starting')
  @_setPristine()
  @scriptEditorResource.apply(@testScript).then (scheduleResponse) =>
    scheduleResponse.operation().then (operation) =>
      @operation = operation
      @_subscribeToOperation(operation)
      return

Controller::createSshKey = (variable) ->
  # TODO: needs more discussion

Controller::save = () ->
  @_updateTestScript()
  @scriptEditorResource.save(@testScript).then (script) =>
    task = {
      subject: {
        scriptUuid: @script.subject.uuid
        environmentUuid: @testScript.environment.uuid
        description: 'autogenerated by script'
        name: 'autogenerated script'
      }
    }
    @taskResource.save(task)
      .catch () =>
        # NOTE: deals with uniqueness errors from the backend
        return angular.noop()
      .then () =>
        @flash.success = @$translate.instant("forms.script.flashes.create.success", @script.subject)
        @$state.go("script", {projectUuid: @project.subject.uuid, scriptUuid: @script.subject.uuid}, {reload: true})
        return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.script.flashes.create.fail", @script.subject)
    @$q.reject(reason)

Controller::delete = () ->
  @scriptResource.delete(@script.subject.uuid).then (script) =>
    @flash.success = @$translate.instant("forms.script.flashes.delete.success", script.subject)
    @$state.go("scripts", {projectUuid: @project.subject.uuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.script.flashes.delete.fail", @script.subject)
    @$q.reject(reason)

angular.module('harrowApp').controller 'scriptEditCtrl', Controller
