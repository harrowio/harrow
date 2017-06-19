app = angular.module("harrowApp")

EnvironmentListCtrl = (
  @project
  @environments
  @secrets
  @environmentResource
  @$translate
  @$state
  @$scope
  @flash
) ->
  @$scope.$on "reloadEnvironments", =>
    @project.environments().then (@environments) =>
  @

EnvironmentListCtrl::secretsFor = (environment) ->

EnvironmentListCtrl::delete = (environment) ->
  if confirm(@$translate.instant("prompts.really?"))
    @environmentResource.delete(environment.subject.uuid).then =>
      @$state.go("projects/edit.environments", {projectUuid: environment.subject.projectUuid}, { reload: true })
      @flash.success = @$translate.instant("environmentList.flashes.deletion.success", environment.subject)
      return

app.controller("environmentListCtrl", EnvironmentListCtrl)
