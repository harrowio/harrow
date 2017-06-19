angular.module('harrowApp').directive 'uiHasView', ($state, $animate, $interpolate) ->
  {
    restrict: 'A'
    link: ($scope, $element, $attrs) ->
      $scope.$watch () =>
        $state.current.data.showViews
      , (views) =>
        if angular.isArray(views)
          if views.indexOf($attrs.uiView) >= 0
            $element.show()
          else
            $element.hide()
  }
