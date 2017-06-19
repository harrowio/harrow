app = angular.module("harrowApp")

EnvironmentFormCtrl = (
  @organization
  @project
  @environment
  @secrets
  @flash
  @$state
  @$timeout
  @$translate
  @environmentResource
  @secretResource
  @$q
  @$rootScope
  @$scope
  @ws
  @modal
) ->
  @envName = @environment.subject.name
  # can't bind to (k,v) of a JS object in ng-repeat
  @variables = []
  if @environment
    # load existing values into array
    angular.forEach @environment.subject.variables, (v, k) =>
      @variables.push
        name: k
        value: v
  @cids = []
  @sshSecrets = @secretsByType(@secrets, "ssh")
  @envSecrets = @secretsByType(@secrets, "env")
  @subscribeSecrets()

  @$scope.$on "$destroy", =>
    @unsubscribeSecrets()

  @showPub = {}

  @

EnvironmentFormCtrl::save = () ->
  @environment.subject.name = @envName
  @saveEnvironment().then (environment) =>
    @$state.go("environment.edit", {uuid: environment.subject.uuid, projectUuid: environment.subject.projectUuid})
    return

EnvironmentFormCtrl::addVariable = () ->
  existingVariable = @variables.filter((v) => v.name == @variableName)[0]
  if existingVariable
    existingVariable.value = @variableValue
  else
    @variables.push
      name: @variableName
      value: @variableValue

  @variableName = ""
  @variableValue = ""

  # save the env so variables look like they work like secrets and ssh keys
  # (immediately saved)
  @saveEnvironment()

EnvironmentFormCtrl::deleteVariable = (variable) ->
  # save the env to be in line with the other forms
  @variables.splice @variables.indexOf(variable), 1
  @saveEnvironment()

EnvironmentFormCtrl::saveEnvironment = () ->
  # copy variables into environment
  @environment.subject.variables = {}
  angular.forEach @variables, (variable) =>
    @environment.subject.variables[variable.name] = variable.value
  @environmentResource.save(@environment).then (environment) =>
    @flash.success = @$translate.instant("forms.environmentForm.flashes.success", environment.subject)
    environment
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.environmentForm.flashes.fail", @environment.subject)
    @$q.reject(reason)


EnvironmentFormCtrl::showSshPublicKeyModal = (secret) ->
  @currentSecret = secret

EnvironmentFormCtrl::showVariableValueModal = (name, value) ->
  @sshKeyDialog = @modal.show(
    title: name,
    content: value,
    name: "Close"
    mode: 'always',
    modal: 'showEnvironmentVariable',
    templateFn: (config) ->
      """
      <div modal="#{config.modal}" modal-mode="#{config.mode}" modal-full-screen>
        <h2 style="margin-bottom: 1em">#{config.title}</h2>
        <p class="variables">
          <code class="variable" style="margin-bottom: 1em">#{config.content}</pre>
        </p>
        <a href="" ng-click="$scope.emit('modal:dismiss')" class="btn btn--border">#{config.name}</a>
      </div>
      """
  )

EnvironmentFormCtrl::addEnvSecret = () ->
  secret =
    subject:
      type: "env"
      name: @secretName
      value: @secretValue
      environmentUuid: @environment.subject.uuid
  @secretResource.save(secret)
    .then () =>
      @secretName = ""
      @secretValue = ""
      @flash.success = @$translate.instant("forms.envForm.flashes.success", secret.subject)
      @reloadSecrets()
    .catch () =>
      @flash.error = @$translate.instant("forms.envForm.flashes.fail", secret.subject)

EnvironmentFormCtrl::addSshSecret = () ->
  secret =
    subject:
      name: @keyName
      environmentUuid: @environment.subject.uuid
      type: "ssh"
  @secretResource.save(secret)
    .then () =>
      @flash.success = @$translate.instant("forms.sshForm.flashes.success", secret.subject)
      @keyName = ""
      @reloadSecrets()
      @subscribeSecret(secret)
    .catch () =>
      @flash.error = @$translate.instant("forms.sshForm.flashes.fail", secret.subject)

EnvironmentFormCtrl::deleteSecret = (secret) ->
  @unsubscribeSecrets()
  if confirm(@$translate.instant("prompts.really?"))
    @secretResource.delete(secret.subject.uuid).then () =>
      @reloadSecrets()

# subscribe to all secrets and reload them on changes
# - their state might be pending and changing while the user is looking at it
EnvironmentFormCtrl::subscribeSecret = (secret) ->
  @cids.push @ws.subRow "secrets", secret.subject.uuid, () =>
    @secretResource.find(secret.subject.uuid).then (secret) =>
      @secrets[idx] = secret
      @$timeout =>
        @$scope.$apply()

EnvironmentFormCtrl::subscribeSecrets = () ->
  angular.forEach @secrets, (secret, idx) =>
    @cids.push @ws.subRow "secrets", secret.subject.uuid, () =>
      @secretResource.find(secret.subject.uuid).then (secret) =>
        @secrets[idx] = secret
        @$timeout =>
          @$scope.$apply()

# unsubscribe from all secrets
EnvironmentFormCtrl::unsubscribeSecrets = () ->
  angular.forEach @cids, (cid) =>
    @ws.unsubscribe cid

EnvironmentFormCtrl::reloadSecrets = () ->
  @environment.secrets()
    .then (secrets) =>
      @unsubscribeSecrets()
      @secrets = secrets
      @sshSecrets = @secretsByType(secrets, "ssh")
      @envSecrets = @secretsByType(secrets, "env")
      @subscribeSecrets()

# filter out secrets that are not type=ssh
EnvironmentFormCtrl::secretsByType = (secrets, type) ->
  s = []
  angular.forEach secrets, (secret) =>
    s.push(secret) if @isType(secret, type)
  s

EnvironmentFormCtrl::isType = (secret, type) ->
  secret.subject.type == type

EnvironmentFormCtrl::delete = () ->
  @environmentResource.delete(@environment.subject.uuid).then =>
    @flash.success = @$translate.instant("environmentList.flashes.deletion.success", @environment.subject)
    @$state.go("environments", {projectUuid: @environment.subject.projectUuid}, { reload: true })
    return

app.controller("environmentFormCtrl", EnvironmentFormCtrl)
