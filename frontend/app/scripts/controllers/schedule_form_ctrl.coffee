app = angular.module("harrowApp")

ScheduleFormCtrl = (
  @organization
  @project
  @script
  @environment
  @task
  @schedule
  @flash
  @$state
  @$translate
  @scheduleResource
  @$q
) ->
  @scheduleType = 'timespec'
  @

ScheduleFormCtrl::save = () ->
  # prevent a situation where both timespec and cronspec are set at the same time
  # clone the object to not clobber the form's state
  schedule = $.extend true, {}, @schedule
  switch @scheduleType
    when "cronspec"
      delete schedule.subject.timespec
    when "timespec"
      delete schedule.subject.cronspec
  @scheduleResource.save(schedule).then (schedule) =>
    @flash.success = @$translate.instant("forms.scheduleForm.flashes.success", schedule.subject)
    @$state.go("tasks/show", {uuid: schedule.subject.taskUuid})
    return
  .catch (reason) =>
    @flash.error = @$translate.instant("forms.scheduleForm.flashes.fail", schedule.subject)
    @$q.reject(reason)

app.controller("scheduleFormCtrl", ScheduleFormCtrl)
