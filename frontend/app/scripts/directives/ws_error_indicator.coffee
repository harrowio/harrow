app = angular.module("harrowApp")

app.directive "wsErrorIndicator", ($rootScope, $timeout, $compile) ->
  {
    restrict: "E"
    templateUrl: 'views/directives/ws_error_indicator.html'
    replace: true
    link: (scope, elem, attrs) ->

      scope.hasError = false

      $rootScope.$on "wsError", ->
        scope.hasError = true
  }
