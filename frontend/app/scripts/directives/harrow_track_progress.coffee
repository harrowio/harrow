Directive = () ->
  template: """<div class="trackProgress">
    <span class="trackProgress__rail">
      <span class="trackProgress__train" ng-style="{left: left, width: width}"></span>
    </span>
    <span class="trackProgress__station" ng-repeat="station in stations track by $index" ng-class="{'active': station.active, 'completed': station.completed}"></span>
  </div>"""
  restrict: 'A'
  scope:
    progress: '@trackProgress'
  replace: true
  link: ($scope, $element, $attrs) ->
    stations = parseInt($attrs.trackProgressStations, 10) || 0

    if stations < 2
      stations = 2
    i = 0
    stations = stations - 1 # zero bound stations
    $scope.stations = []
    while i <= stations
      station =
        active: false
        completed: false
      $scope.stations.push station
      i++

    $scope.$watch 'progress', (progress) ->
      progress = parseFloat(progress) || 0
      $scope.left = (100 * progress) - 50 + "%"
      $scope.width = 100 / ($scope.stations.length - 1) + "%"
      #
      if progress > 1
        $scope.left = "100%"

      $scope.stations.forEach (station, i) ->
        stationPoint = (i / ($scope.stations.length - 1))
        station.active = false
        station.completed = false
        if i == 0 || stationPoint <= progress
          station.active = true

        if stationPoint <= ( progress - (1 / ($scope.stations.length - 1)))
          station.completed = true


angular.module('harrowApp').directive 'trackProgress', Directive
