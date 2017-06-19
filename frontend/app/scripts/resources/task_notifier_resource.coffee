app = angular.module("harrowApp")

app.factory "TaskNotifier", ($injector, $http) ->
  TaskNotifier = (data) ->
    $.extend(true, @, data)
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  TaskNotifier

app.factory "taskNotifierResource", (Resource, TaskNotifier) ->
  TaskNotifierResource = () ->
    Resource.call(@)
    @

  TaskNotifierResource:: = Object.create(Resource::)
  TaskNotifierResource::basepath = "/job-notifiers"
  TaskNotifierResource::model = TaskNotifier

  TaskNotifierResource::_save = TaskNotifierResource::save

  TaskNotifierResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new TaskNotifierResource()
