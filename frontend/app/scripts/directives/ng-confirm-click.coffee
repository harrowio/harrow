angular.module('harrowApp').directive 'ngConfirmClick', (
  $window
) ->
  {
    priority: -1
    restrict: 'A'
    link: ($scope, $element, $attrs) ->
      message = $attrs.ngConfirmClick || 'Are you sure?'
      clickAction = $attrs.ngClick

      $element.bind 'click', (e) ->
        e.stopImmediatePropagation()
        e.preventDefault()
        if $window.confirm(message)
          $scope.$eval(clickAction)

  }
