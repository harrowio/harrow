Directive = () ->
  restrict: 'A'
  template: '<span svg-icon="{{icon}}" svg-icon-size="{{size}}" class="iconColor"></span>'
  scope: {}
  link: ($scope, $element, $attrs) ->
    $attrs.$observe 'svgIconSize', (size) ->
      $scope.size = size
    $attrs.$observe 'statusIcon', (status) ->
      icon = 'icon-info'
      if status == 'success'
        icon = 'icon-complete'
      else if status == 'timeout'
        icon = 'icon-clock'
      else if status == 'canceled'
        icon = 'icon-canceled'
      else if status == 'active' || status == 'pending' || status == 'running'
        icon = 'icon-spinner'
      else if status == 'failure' || status == 'failed' || status == 'fatal' || status == 'fatalerr'
        icon = 'icon-error'

      $scope.icon = icon

angular.module('harrowApp').directive 'statusIcon', Directive
