app = angular.module("harrowApp")

ValidateMatch = () ->
  restrict: "A"
  link: ($scope, $element, $attrs) ->
    ngModel = $element.controller("ngModel")
    $scope.$watch ->
      $attrs.validateMatch
    , (v) ->
      ngModel.$validate()

    ngModel.$validators.match = (modelValue) ->
      # initially modelValue is undefined, so it never matches the
      # empty string
      if modelValue == undefined
        modelValue = ""

      modelValue == $attrs.validateMatch

app.directive("validateMatch", ValidateMatch)
