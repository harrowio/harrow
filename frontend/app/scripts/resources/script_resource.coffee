app = angular.module("harrowApp")

app.factory "Script", ($injector) ->
  Script = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    if @subject
      previewLines = @subject.body.split("\n").slice(0,10)
      if previewLines.length == 10
        previewLines[9] = "…"
      @preview = previewLines.join("\n")
    @

  Script::project = () ->
    @projectResource.fetch(@_links.project.href)

  Script

app.factory "scriptResource", (Resource, Script) ->
  ScriptResource = () ->
    Resource.call(@)
    @

  ScriptResource:: = Object.create(Resource::)
  ScriptResource::basepath = "/tasks"
  ScriptResource::model = Script

  new ScriptResource()
