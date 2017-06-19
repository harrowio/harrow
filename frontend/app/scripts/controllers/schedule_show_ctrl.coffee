app = angular.module("harrowApp")

ScheduleShowCtrl = (
  @$scope
  @$state
  $timeout
  @$translate
  @flash
  @task
  @taskName
  @script
  @repositories
  @operations
  @schedule
  @operationResource
  @scheduledExecutions
  @organization
  @project
  @ws
  initEvents
) ->
  @redirectRunNow()
  @

ScheduleShowCtrl::redirectRunNow = () ->
  if @isRunNowSchedule() and @operations.length > 0
    @$state.go('operations/show', {uuid: @operations[0].subject.uuid})

ScheduleShowCtrl::isRunNowSchedule = () ->
  timespec = @schedule.subject.timespec
  timespec == "now"

app.controller("scheduleShowCtrl", ScheduleShowCtrl)
