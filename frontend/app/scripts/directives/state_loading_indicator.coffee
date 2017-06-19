app = angular.module("harrowApp")

app.directive 'stateLoadingIndicator', ($rootScope, $timeout) ->
  {
    restrict: 'E'
    template: '<div class=\'loading-indicator\' ng-show=\'isStateLoading\'>' + '<div class=\'loading-indicator-body\'>' + '<h3 class=\'loading-title\'>Loading...</h3>' + '</div>' + '</div>'
    replace: true
    link: (scope, elem, attrs) ->
      timeout = null
      scope.isStateLoading = false
      $rootScope.$on '$stateChangeStart', ->
        $timeout.cancel timeout
        # prevent indicator when a ws error happened
        unless scope.wsError
          timeout = $timeout( ->
            scope.isStateLoading = true
          , 100)
      cancel = ->
        $timeout.cancel timeout
        scope.isStateLoading = false
      $rootScope.$on '$stateChangeSuccess', cancel
      $rootScope.$on "$stateChangeError", cancel

      # turn off state loading indicator when a ws error happened to prevent
      # them from overlapping
      $rootScope.$on "wsError", ->
        scope.isStateLoading = false
        scope.wsError = true
  }
