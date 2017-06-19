svg = require('../../../bower_components/style-guide/source/images/icons.svg')
angular.module('harrowApp').directive 'svgIcon', (
  $document
  $http
  $q
) ->
  deferredPromises = []
  readyPromise = $q.defer()
  $http.get(svg).success (svg) ->
    el = document.createElement('div')
    el.style.display = 'none'
    el.setAttribute('id', 'svg-icons')
    el.innerHTML = svg
    $document.find('body').append(el)
    readyPromise.resolve()
  {
    restrict: 'A',
    template: '<svg xmlns:xlink="http://www.w3.org/1999/xlink" version="1.1" viewBox="0 0 36 36"></svg>'
    replace: true
    link: ($scope, $element, $attrs) ->
      $element.attr('width', 36)
      $element.attr('height', 36)
      $scope.$watch $attrs.svgIcon, (thing) ->
        readyPromise.promise.then () ->
          symbol = document.querySelector("#svg-icons ##{$attrs.svgIcon}")
          if symbol
            $element[0].setAttribute('viewBox', symbol.getAttribute('viewBox'))
            $element.html(symbol.innerHTML)

      $attrs.$observe 'width', (size) ->
        $element.attr('width', size) if size

      $attrs.$observe 'height', (size) ->
        $element.attr('height', size) if size

      $attrs.$observe 'svgIconSize', (size) ->
        if size
          $element.attr('width', size)
          $element.attr('height', size)

  }
