Directive = () ->
  restrict: 'A'
  require: 'ngModel'
  link: ($scope, $element, $attrs, $ngModel) ->
    read = ->
      if $element.html() != ($ngModel.$viewValue || '')
        $ngModel.$setViewValue $element.html()

    $ngModel.$render = ->
      $element.html($ngModel.$viewValue || '')

    if $attrs.hasOwnProperty('alphaNumericOnly')
      $element.bind 'keypress', (e) ->
        # disable everything except `a-zA-Z0-9_-`
        if !(e.which != 8 && e.which != 0 && (e.which == 45 || e.which == 95 || (e.which >= 48 && e.which <= 57) || (e.which >= 65 && e.which <= 90) || (e.which >= 97 && e.which <= 122) ))
          e.preventDefault()
    $element.bind 'blur', ->
      $scope.$apply(read)

angular.module('harrowApp').directive 'contenteditable', Directive
