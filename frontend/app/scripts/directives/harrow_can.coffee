app = angular.module("harrowApp")

app.directive "harrowCan", ($rootScope, $window, $parse, $filter) ->
  {
    restrict: "AE"
    link: ($scope, $element, $attrs) ->
      can = $attrs.harrowCan
      disable = $attrs.hasOwnProperty('canDisable') || false
      if $attrs.canAction
        can = $attrs.canAction
      [action, relname] = can.split '-'
      relname = "self" unless relname
      subject = $parse($attrs.canSubject)($scope)
      relname = $filter('dashCase')(relname)
      rel = if subject?._links then subject._links[relname] else undefined
      disableClick = (e) ->
        e.preventDefault()
        return false
      if rel? && rel[action]?
        $($element).show()
        $element.unbind 'click', disableClick
        $element.css 'cursor',''
      else
        if disable
          $element.bind 'click', disableClick
          $element.css 'cursor', 'not-allowed'
        else
          $($element).hide()
  }
