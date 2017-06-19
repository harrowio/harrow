app = angular.module('harrowApp')

app.factory 'Stencil', ($injector) ->
  Stencil = (data) ->
    $.extend(true, @, data)
    @
  Stencil

app.factory 'stencilResource', (Resource,Stencil) ->
  StencilResource = () ->
    Resource.call(@)
    @

  StencilResource:: = Object.create(Resource::)
  StencilResource::basepath = "/stencils"
  StencilResource::model = Stencil
  new StencilResource()
