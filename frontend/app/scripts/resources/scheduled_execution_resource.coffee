app = angular.module("harrowApp")

app.factory "ScheduledExecution", () ->
  ScheduledExecution = (data) ->
    $.extend(true, @, data)
    @

  ScheduledExecution

app.factory "scheduledExecutionResource", (Resource, ScheduledExecution) ->
  ScheduledExecutionResource = () ->
    Resource.call(@)
    @

  ScheduledExecutionResource:: = Object.create(Resource::)
  # TODO: bogus, cannot be fetch apart from as a subresource on tasks
  ScheduledExecutionResource::basepath = "/scheduled-executions"
  ScheduledExecutionResource::model = ScheduledExecution

  ScheduledExecutionResource::_save = ScheduledExecutionResource::save

  ScheduledExecutionResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new ScheduledExecutionResource()
