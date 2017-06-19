ace = require 'brace'
require 'brace/mode/sh'
require 'brace/theme/solarized_light'

Directive = () ->
  scope:
    text: '=text'
  link: (
    $scope
    $element
    $attrs
  ) ->
    editor = ace.edit($element[0])
    session = editor.getSession()
    session.setMode("ace/mode/#{$attrs.mode}") if $attrs.mode
    editor.setTheme("ace/theme/#{$attrs.theme}") if $attrs.theme

    oldValue = $scope.text
    if $scope.text
      editor.setValue($scope.text)
      editor.clearSelection()
      editor.focus()

    editor.on 'change', () ->
      newValue = session.getValue()
      return if oldValue == newValue
      unless angular.isUndefined(oldValue)
        $scope.$emit('uiAce:textChanged', newValue)
      oldValue = newValue

angular.module('harrowApp').directive 'uiAce', Directive
