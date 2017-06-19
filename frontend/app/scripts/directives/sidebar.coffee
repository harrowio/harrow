angular.module('harrowApp').directive 'sidebar', ->
  {
    restrict: 'C',
    link: ($scope, $element, $attrs) ->
      angular.element('body').bind 'click', (event) ->
        if angular.element(event.target).closest('.sidebar__header__account').length == 0 && angular.element('.sidebar__content').hasClass('sidebar__content--second')
          angular.element('.sidebar__content').removeClass('sidebar__content--second')
      $element.find('.sidebar__header__account').bind 'click', ->
        $element.find('.sidebar__content').toggleClass('sidebar__content--second')
  }
