app = angular.module("harrowApp")

app.factory "Environment", ($injector) ->
  Environment = (data) ->
    $.extend(true, @, data)
    @projectResource = $injector.get("projectResource")
    @secretResource = $injector.get("secretResource")
    @

  Environment::project = () ->
    @projectResource.fetch(@_links.project.href)

  Environment::secrets = () ->
    @secretResource.fetch(@_links.secrets.href)

  Environment

app.factory "environmentResource", (Resource, Environment) ->
  EnvironmentResource = () ->
    Resource.call(@)
    @

  EnvironmentResource:: = Object.create(Resource::)
  EnvironmentResource::basepath = "/environments"
  EnvironmentResource::model = Environment

  new EnvironmentResource()
