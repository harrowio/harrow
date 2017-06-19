###*
* @ngdoc directive
* @name sidebarButton
* @restrict A
* @description Toggles sidebar--open, and removes sidebar--open when clicked outside element
###

angular.module('harrowApp').directive 'sidebarToggle', () ->
  {
    restrict: 'A'
    link: ($scope, $element) ->
      angular.element('body').bind 'click', (event) ->
        if angular.element(event.target).closest('.sidebar').length == 0 && angular.element(event.target).closest('.sidebar-toggle').length == 0 && angular.element('.sidebar').hasClass('sidebar--open')
          angular.element('.sidebar').removeClass('sidebar--open')

      angular.element('.sidebar a').bind 'click', ->
        angular.element('.sidebar').removeClass('sidebar--open')

      $element.bind 'click', (event) ->
        event.stopPropagation()
        angular.element('.sidebar').toggleClass('sidebar--open')
  }
