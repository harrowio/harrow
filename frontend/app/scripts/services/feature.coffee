Feature = (
  @$q
  @appFeatureResource
) ->
  @enabledFeatures = []
  @

Feature::isEnabled = (expectedFeature) ->
  if !angular.isArray(@enabledFeatures) || !angular.isString(expectedFeature) || expectedFeature.length == 0
    return true

  @enabledFeatures.filter (feature) ->
    feature.subject.name == expectedFeature
  .some (feature) ->
    if feature.subject.hasOwnProperty('enableAt')
      moment().isAfter(feature.subject.enableAt)
    else
      feature.subject.enabled

Feature::isDisabled = (value) ->
  !@isEnabled(value)

Feature::loadFeatures = () ->
  @appFeatureResource.all().then (features) =>
    @enabledFeatures = features
  .catch =>
    @$q.when()

angular.module('harrowApp').service 'feature', Feature

angular.module("harrowApp").factory 'AppFeature', () ->
  AppFeature = (data) ->
    $.extend(true, @, data)
    @

  AppFeature

angular.module("harrowApp").factory 'appFeatureResource', (
  Resource
  AppFeature
) ->
  AppFeatureResource = () ->
    Resource.call(@)
    @
  AppFeatureResource:: = Object.create(Resource::)
  AppFeatureResource::model = AppFeature
  AppFeatureResource::basepath = '/api-features'
  new AppFeatureResource()
