app = angular.module("harrowApp")

app.factory "ScriptCard", ($injector, Operation, Environment) ->
  ScriptCard = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")

    if data.subject.lastOperation
      @subject.lastOperation = new Operation({subject:@subject.lastOperation})

    @subject.enabledEnvironments.forEach (env, index) ->
      env = new Environment({subject:env})
    @

  ScriptCard::project = () ->
    @projectResource.fetch(@_links.project.href)

  ScriptCard

app.factory "scriptCardResource", (Resource, ScriptCard) ->
  ScriptResource = () ->
    Resource.call(@)
    @

  ScriptResource:: = Object.create(Resource::)
  ScriptResource::basepath = "/scripts"
  ScriptResource::model = ScriptCard

  new ScriptResource()
