app = angular.module("harrowApp")

Controller = (
  $scope
  $attrs
  @scheduleResource
) ->

  @showTo = $attrs.showTo || 10

  @

OperationList = () ->
  restrict: "E"
  scope:
    operations: "="
  controller: Controller
  controllerAs: "operationList"
  templateUrl: 'views/directives/operation_list.html'

app.directive("operationList", OperationList)
