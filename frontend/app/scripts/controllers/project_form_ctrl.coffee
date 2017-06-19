app = angular.module("harrowApp")

ProjectFormCtrl = (
  @organization
  @project
  @members
  @projectResource
  @flash
  @$scope
  @$state
  @$translate
  @$q
) ->
  @$scope.$on "tasksChanged", =>
    @$scope.$broadcast "reloadTasks"
  @$scope.$on "environmentsChanged", =>
    @$scope.$broadcast "reloadEnvironments"
  @

ProjectFormCtrl::save = () ->
  @projectResource.save(@project).then (project) =>
    @flash.success = @$translate.instant("forms.projectForm.flashes.success", project.subject)
    unless @$state.includes("projects/edit")
      @$state.go("projects/edit", {uuid: project.subject.uuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.projectForm.flashes.fail", @project.subject)
    @$q.reject(reason)

app.controller("projectFormCtrl", ProjectFormCtrl)
