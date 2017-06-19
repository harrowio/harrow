app = angular.module("harrowApp")

app.factory "ScriptEditor", ($injector, $q, $log, $timeout, operationResource, taskResource) ->
  Script = (data) ->
    $.extend(true, @, data)
    @_pollOperationsTimeout = null
    @

  Script::_pollOperations = (taskResponse, attempts = 0) ->
    operationResource.fetch(taskResponse._links.operations.href).then (response) =>
      operation = response.find (operation) =>
        operation.subject.parameters.scheduleUuid == @subject.uuid
      if operation
        $timeout.cancel(@_pollOperationsTimeout)
        operation
      else if attempts < 10
        @_pollOperationsTimeout = $timeout =>
          @_pollOperations(taskResponse, attempts + 1)
        , 500
      else
        $log.error('Could not get operation after 5 seconds')
        $q.reject()

  Script::operation = () ->
    taskResource.fetch(@_links.job.href).then (taskResponse) =>
      @_pollOperations(taskResponse)

  Script

app.factory "scriptEditorResource", ($http, endpoint, Resource, ScriptEditor) ->
  ScriptResource = () ->
    Resource.call(@)
    @

  ScriptResource:: = Object.create(Resource::)
  ScriptResource::basepath = "/script-editor"
  ScriptResource::model = ScriptEditor

  ScriptResource::_makeRequest = (url, input) ->
    obj = angular.copy(input)
    obj.task = input.script
    obj.taskUuid = input.scriptUuid
    obj.jobUuid = input.taskUuid
    $http.post(url, obj).then (response) =>
      @makeModel(response.data)

  ScriptResource::apply = (obj) ->
    url = "#{endpoint}#{@basepath}/apply"
    @_makeRequest(url, obj)

  ScriptResource::diff = (obj) ->
    url = "#{endpoint}#{@basepath}/diff"
    @_makeRequest(url, obj)

  ScriptResource::save = (obj) ->
    url = "#{endpoint}#{@basepath}/save"
    @_makeRequest(url, obj)

  new ScriptResource()
