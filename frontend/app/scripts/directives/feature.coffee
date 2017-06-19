###*
# @ngdoc directive
# @name feature
# @restrict AE
# @description Shows view blocks when a feature is enabled.
# If a given key is null or undefined, the block will show by default.
# @param {string} feature key to check if it is enabled
# @example
<example>
  <file name="index.html">
    <a href="/oauth" feature="oauth">Login Via OAuth</a>
  </file>
  <file name="api_app_feautres.json">
    {
      "collection": [
        {
          "subject": {
            "name": "oauth",
            "enabled": true
          }
        }
      ]
    }
  </file>
</example>
###
FeatureController = ($scope, $element, $attrs, feature) ->
  if feature.isEnabled($scope.feature)
    $element.removeClass('ng-hide')
  else
    $element.addClass('ng-hide')
  @

FeatureDirective = () ->
  restrict: 'AE'
  scope:
    feature: '@feature'
  controller: 'featureDirectiveCtrl'

angular.module('harrowApp').directive 'feature', FeatureDirective
angular.module('harrowApp').controller 'featureDirectiveCtrl', FeatureController

###*
# @ngdoc directive
# @name noFeature
# @restrict AE
# @description Shows view blocks when a feature is disabled.
# If a given key is null or undefined, the block will hide by default.
# @param {string} feature key to check if it is disabled
# @example
<example>
  <file name="index.html">
    <a href="/standard" no-feature="oauth">Login Standard</a>
  </file>
  <file name="api_app_feautres.json">
    {
      "collection": [
        {
          "subject": {
            "name": "oauth",
            "enabled": false
          }
        }
      ]
    }
  </file>
</example>
###
NoFeatureController = ($scope, $element, $attrs, feature) ->
  if feature.isDisabled($scope.feature)
    $element.removeClass('ng-hide')
  else
    $element.addClass('ng-hide')
  @

NoFeatureDirective = () ->
  restrict: 'AE'
  scope:
    feature: '@noFeature'
  controller: 'noFeatureDirectiveCtrl'

angular.module('harrowApp').controller 'noFeatureDirectiveCtrl', NoFeatureController
angular.module('harrowApp').directive 'noFeature', NoFeatureDirective
