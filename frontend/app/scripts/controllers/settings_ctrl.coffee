app = angular.module("harrowApp")

SettingsCtrl = (
  @$scope
  @$state
  @oauth
  @flash
  @$translate
  @authentication
  @userResource
  @$http
  @endpoint
) ->
  @refreshGithubStatus()
  @checkMFA()
  @user = @authentication.currentUser

  @


SettingsCtrl::refreshGithubStatus = ->
  @oauth.pingGithub().then (data) =>
    @githubAccessible = data.data.status == "up"
  .catch (reason) =>
    @githubAccessible = false

SettingsCtrl::checkMFA = ->
  @checkTotp()

SettingsCtrl::checkTotp = ->
  @totpEmail  = @authentication.currentUser.subject.email
  @totpEnabled = @authentication.currentUser.totpEnabled()

SettingsCtrl::enableTotp = ->
  @$state.go("enable_totp") unless @totpEnabled

SettingsCtrl::disableTotp = ->
  @$state.go("disable_totp") if @totpEnabled

SettingsCtrl::connectGithub = ->
  @oauth.authorizeGithub()

SettingsCtrl::updatePersonalSettings = ->
  @userResource.save(@user).then () =>
    @flash.success = @$translate.instant("forms.userSettings.flashes.success")
    @$scope.$emit('modal:dismiss')
    return
  .catch () =>
    @flash.error = @$translate.instant("forms.userSettings.flashes.failure")
    return

SettingsCtrl::disconnectGithub = ->
  @oauth.deauthorizeGithub().then () =>
    @refreshGithubStatus()

SettingsCtrl::resetPrompts = ->
  url = @endpoint + "/prompts/all"
  @$http.delete url

SettingsCtrl::leaveProject = (project) ->
  return unless confirm(@$translate.instant("prompts.really?"))
  project.leave().then () =>
    @projects = @scope.$resolve.projects.filter (existingProject) ->
      existingProject.subject.uuid != project.subject.uuid

app.controller("settingsCtrl", SettingsCtrl)
