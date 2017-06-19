ZeroClipboardSWF = String(require('zeroclipboard/dist/ZeroClipboard.swf'))
ZeroClipboard = require('zeroclipboard/dist/ZeroClipboard')

ZeroClipboard.config
  swfPath: ZeroClipboardSWF
  trustedDomains: ['*']
  allowScriptAccess: 'always'
  forceHandCursor: true

Directive = (
  $compile
) ->
  restrict: 'A'
  scope:
    clipboard: "=clipboard"
  replace: true
  template: '<div><a href="" clip-copy data-clipboard-text="{{clipboard}}" title="Click to copy to clipboard">Copy to Clipboard</a></div>'
  link: ($scope, $element, $attrs) ->
    handler = new ZeroClipboard($element)
    handler.on 'ready', ->
      handler.on 'copy', (event) ->
        event.clipboardData.setData 'text/plain', $scope.clipboard
        false
      handler.on 'aftercopy', (event) ->
        $element.find('a').html('<span svg-icon="icon-complete" svg-icon-size="12" class="iconColor"></span> Copied to clipboard')
        $compile($element.find('a'))($scope)
        $element.blur()
        false

angular.module('harrowApp').directive 'clipboard', Directive
