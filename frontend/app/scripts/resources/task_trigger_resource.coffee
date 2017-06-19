app = angular.module("harrowApp")

app.factory "taskTrigger", ($injector) ->
  taskTrigger = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @$http = $injector.get("$http")
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  taskTrigger::project = () ->
    @projectResource.fetch(@_links.project.href)

  taskTrigger

app.factory "taskTriggerResource", (Resource, taskTrigger) ->
  taskTriggerResource = () ->
    Resource.call(@)
    @

  taskTriggerResource:: = Object.create(Resource::)
  taskTriggerResource::basepath = "/job-triggers"
  taskTriggerResource::model = taskTrigger

  taskTriggerResource::_save = taskTriggerResource::save

  taskTriggerResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new taskTriggerResource()
