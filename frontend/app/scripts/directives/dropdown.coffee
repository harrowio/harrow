app = angular.module("harrowApp")

Dropdown = (@$scope) ->

  @visible = false

  d = $(document)
  handler = (e) =>
    @maybeHide(e)
  d.on "click", handler
  $scope.$on "$destroy", ->
    d.off "click", handler
  @

Dropdown::maybeHide = (e) ->
  if $(e.target).parents("span.dropdown-wrapper").length == 0
    @hide()
    @$scope.$digest()
  return true

Dropdown::toggle = ->
  @visible = !@visible

Dropdown::hide = ->
  @visible = false

app.directive "dropdown", ($rootScope) ->
  {
    restrict: "E"
    controller: Dropdown
    controllerAs: "dropdown"
    templateUrl: 'views/directives/dropdown.html'
    transclude: true
  }
