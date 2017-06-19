app = angular.module("harrowApp")

app.directive "highlight", ($rootScope, $window, @ws) ->
  {
    restrict: "A"
    link: (scope, elem) ->
      scope.$watch () ->
        elem.innerHTML
      , () ->
        hljs.highlightBlock(elem[0])
  }
