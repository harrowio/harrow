Directive = () ->
  restrict: 'A'
  template: """<div>
    <small ng-repeat="rule in rules">
      <span ng-if="$last && !$first"> and </span>
      {{'forms.notifiers.taskNotifier.options.triggerAction.' + rule.subject.matchActivity | translate }}<span ng-if="!$last">, </span>
    </small>
  </div>"""
  replace: true
  scope:
    rules: '=notificationRules'
  link: ($scope, $element) ->
    $scope.$watchCollection 'rules', (rules) ->


angular.module('harrowApp').directive 'notificationRules', Directive
