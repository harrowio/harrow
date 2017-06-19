Directive = ($http, endpoint) ->
  restrict: 'A'
  templateUrl: 'views/directives/modal.html'
  transclude: true
  scope: false
  replace: true

  link: ($scope, $element, $attrs) ->
    url = "#{endpoint}/prompts/#{$attrs.modal}"
    $http.get(url).success (data) ->
      $element.remove() if $attrs.modalMode != 'always' && data?.dismissed
    $scope.isFullScreen = false

    handleDismiss = (event) ->
      if event.target && !angular.element(event.target).hasClass('modal__dismiss')
        $http.delete(url)
      $element.remove()
      $scope.$emit 'modal:dismissed'

    $element.find('a').bind 'click', handleDismiss
    $scope.$on 'modal:dismiss', handleDismiss

    $scope.mode = $attrs.modalMode
    if $attrs.hasOwnProperty('modalSticky')
      $element.find('.modal__dismiss').remove()
    if $attrs.hasOwnProperty('modalFullScreen')
      $scope.isFullScreen = true

angular.module('harrowApp').directive 'modal', Directive
