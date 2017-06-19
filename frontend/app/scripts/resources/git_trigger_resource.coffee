app = angular.module("harrowApp")

app.factory "GitTrigger", ($injector) ->
  GitTrigger = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @$http = $injector.get("$http")
    if @subject
      @subject.taskUuid = @subject.jobUuid
    @

  GitTrigger::project = () ->
    @projectResource.fetch(@_links.project.href)

  GitTrigger

app.factory "gitTriggerResource", (Resource, GitTrigger) ->
  GitTriggerResource = () ->
    Resource.call(@)
    @

  GitTriggerResource:: = Object.create(Resource::)
  GitTriggerResource::basepath = "/git-triggers"
  GitTriggerResource::model = GitTrigger

  GitTriggerResource::_save = GitTriggerResource::save

  GitTriggerResource::save = (object) ->
    object.subject.jobUuid = object.subject.taskUuid
    @_save(object)

  new GitTriggerResource()
