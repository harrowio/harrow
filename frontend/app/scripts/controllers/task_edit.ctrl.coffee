Controller = (
  @task
  @taskResource
  @flash
  @$translate
  @$state
) ->
  @


Controller::delete = () ->
  @taskResource.delete(@task.subject.uuid).then () =>
    @flash.success = @$translate.instant('forms.taskForm.flashes.delete.success')
    @$state.go 'projects/show', { projectUuid: @task.subject.projectUuid }
    return
  .catch =>
    @flash.error = @$translate.instant('forms.taskForm.flashes.delete.fail')

angular.module('harrowApp').controller 'taskEditCtrl', Controller
