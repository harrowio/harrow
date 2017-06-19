Directive = () ->
  restrict: 'A'
  template: "<span>Will run {{schedule | formattedSchedule:'calendar'}} <small>(in {{schedule | formattedSchedule:'fromNow'}})</small></span>"
  replace: true
  scope:
    schedule: '=scheduleNextRun'
  link: ($scope, $element) ->

    $scope.$watch 'schedule', (schedule) ->
      if angular.isDefined(schedule.subject.timespec) || angular.isDefined(schedule.subject.cronspec)
        $element.show()
      else
        $element.hide()
    , true
angular.module('harrowApp').directive 'scheduleNextRun', Directive
