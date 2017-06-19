app = angular.module("harrowApp")

app.factory "Loggable", ($injector) ->
  Loggable = (data) ->
    $.extend(true, @, data)
    @

  Loggable

app.factory "loggableResource", (Resource, Loggable) ->
  LoggableResource = () ->
    Resource.call(@)
    @

  LoggableResource:: = Object.create(Resource::)
  LoggableResource::basepath = "/logs"
  LoggableResource::model = Loggable

  new LoggableResource()
