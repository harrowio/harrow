Controller = (
  @project
  @projectResource
  @$translate
  @flash
  @$state
  @menuItems
) ->
  @menu = @menuItems.projectEdit
  @

Controller::save = ->
  @projectResource.save(@project).then () =>
    @flash.success = @$translate.instant('forms.project.flashes.create.success')
    @$state.go 'projects/edit', { projectUuid: @project.subject.uuid }
    return
  .catch =>
    @flash.error = @$translate.instant('forms.project.flashes.create.fail')
    return

Controller::delete = ->
  @projectResource.delete(@project.subject.uuid).then () =>
    @flash.success = @$translate.instant('forms.project.flashes.deletion')
    @$state.go 'organization', { uuid: @project.subject.organizationUuid }, {reload: true}
    return
  .catch =>
    @flash.error = @$translate.instant('forms.project.flashes.deletionFail')
    return

angular.module('harrowApp').controller 'projectCtrl', Controller
