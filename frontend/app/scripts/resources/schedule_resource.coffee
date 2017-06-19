app = angular.module("harrowApp")

app.factory "Schedule", ($injector) ->
  Schedule = (data) ->
    $.extend(true, @, data)
    @scheduledExecutionResource = $injector.get("scheduledExecutionResource")
    @operationResource = $injector.get("operationResource")
    @taskResource = $injector.get("taskResource")
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  Schedule::task = ->
    @taskResource.find(@subject.taskUuid)

  Schedule::scheduledExecutions = ->
    @scheduledExecutionResource.fetch @_links['scheduled-executions'].href

  Schedule::operations = ->
    @operationResource.fetch @_links['operations'].href

  Schedule

app.factory "scheduleResource", (Resource, Schedule) ->
  ScheduleResource = () ->
    Resource.call(@)
    @

  ScheduleResource:: = Object.create(Resource::)
  ScheduleResource::basepath = "/schedules"
  ScheduleResource::model = Schedule

  ScheduleResource::_save = ScheduleResource::save

  ScheduleResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new ScheduleResource()
