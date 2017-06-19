angular.module('harrowApp').directive 'formattedGitTrigger', () ->
  {
    restrict: 'A'
    replace: true
    scope: {
      trigger: '=formattedGitTrigger'
      extra: '=ngInit'
    }
    template: '<span/>'
    link: ($scope, $element, $attrs) ->
      out = []
      out.push 'When'
      out.push $scope.trigger.subject.matchRef
      out.push 'is'
      out.push $scope.trigger.subject.changeType
      out.push 'then trigger'
      if $scope.extra.script
        out.push "<strong>#{$scope.extra.script.subject.name}</strong>"
      if $scope.extra.environment
        out.push 'on'
        out.push "<strong>#{$scope.extra.environment.subject.name}</strong>"
      out = out.join(' ')
      $element.html(out)
  }
