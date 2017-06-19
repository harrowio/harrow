app = angular.module("harrowApp")

app.directive "elapsedFrom", ($rootScope, $interval, $filter) ->
  updateTimes = (scope, el) ->
    duration = moment.duration(moment(scope.elapsedTo).diff(scope.elapsedFrom))
    el.html($filter('momentDurationFormat')(duration))
  {
    restrict: "A"
    scope:
      elapsedFrom: "="
      elapsedTo: "="
    link: (scope, elem) ->

      # ivokeApply=false, skip model dirty checking
      i = $interval () ->
        updateTimes(scope, elem)
      , 1000, 0, false

      scope.$on "$destroy", ->
        $interval.cancel(i)

      scope.$watchGroup ["elapsedFrom", "elapsedTo"], () ->
        updateTimes(scope, elem)
  }
