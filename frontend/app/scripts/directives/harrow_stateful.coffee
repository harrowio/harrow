###*
# @ngdoc directive
# @name harrowStateful
# @restrict A
# @description Transitioning state element for async elements.
# @priority 10
# @param {object} object definiton of content and attributes for transition states
# @example
 <example>
   <file name="index.html">
     <script>
       angular.module('harrowApp', [])
         .controller('ExampleCtrl', function ($scope) {
           $scope.isSaving = false
           $scope.isSaved = false
           $scope.onSave = function () {
             $scope.isSaving = true
             setTimeout(function () {
               $scope.isSaved = true
               $scope.$digest()
             })
           }
           $scope.linkOptions = {
             pending: {
               watch: 'isSaving',
               content: 'Saving'
             },
             completed: {
               watch: 'isSaved',
               content: 'Saved',
               attrs: {
                 ngClick: null, // Disables the click event listener
                 uiSref: 'dashboard'
               }
             }
           }
         })
     </script>
     <a harrow-stateful="linkOptions" ng-click="onSave">Save</a>
   </file>
 </example>
###

unbindEvents = ($element, key) ->
  events = 'click dblclick mousedown mouseup mouseover mouseout mousemove mouseenter mouseleave keydown keyup keypress submit focus blur copy cut paste'.split(' ')
  events.forEach (eventKey) ->
    dashCase = key.replace(/([A-Z])/g, '-$1').toLowerCase()
    if dashCase == "ng-#{eventKey}"
      $element.unbind(eventKey)


setStatus = (options = {}, $element, $attrs, $compile, $scope) ->
  statefulAttr = $element.attr('harrow-stateful')
  $element.removeAttr('harrow-stateful')
  $element.html(options.content) if options.content
  Object.keys(options.attrs || {}).forEach (key) ->
    unbindEvents($element, key)
    $attrs.$set key, options.attrs[key]
  $compile($element)($scope.$parent)
  $element.attr('harrow-stateful', statefulAttr)


angular.module('harrowApp').directive 'harrowStateful', ($parse, $compile) ->
  {
    restrict: 'A'
    scope:
      state: '=harrowStateful'
    link: ($scope, $element, $attrs) ->
      $scope.$watch 'state', (obj) ->
        setStatus(obj, $element, $attrs, $compile, $scope) if obj
      , true
  }
