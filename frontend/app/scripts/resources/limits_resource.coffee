app = angular.module('harrowApp')

app.factory "Limit", ($injector, $http) ->
  Limit = (data) ->
    $.extend(true, @, data)
    @

  Limit

app.factory 'limitResource', (Resource, Limit) ->
  LimitResource = () ->
    Resource.call(@)
    @

  LimitResource:: = Object.create(Resource::)
  LimitResource::basepath = "/limits"
  LimitResource::model = Limit
  new LimitResource()
