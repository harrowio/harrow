Controller = (
  @$scope
  @$element
  @$attrs
  @$translate
  @$transclude
  @$compile
) ->
  @translationRoot = @$scope.harrowForm?.translationRoot
  @group = @$element.parents('.field__group')
  @input = @$transclude().parent().find('input,select,textarea')

  @model = @input.controller('ngModel')
  @input.attr('placeholder', @maybeTranslate("#{@translationRoot}.placeholder.#{@model?.$name}"))
  @label = @maybeTranslate("#{@translationRoot}.label.#{@model?.$name}")
  @type = @input.attr('type')
  @render()

  if @group.find('input,select').length > 1
    @group.find('.field__input:not(:last-child)').addClass('field__input--join')

  @$scope.harrowForm?.addInput(@)
  @

Controller::render = () ->
  transclude = @$transclude().parent().html()
  el = angular.element("""<div><label>#{transclude}<span/> <span ng-bind-html="(harrowFieldInput.label | translate | toTrusted)"/></label></div>""")
  if @type == 'radio'
    el.addClass('field__radio')
  else if @type == 'checkbox'
    el.addClass('field__checkbox')
  else
    el = angular.element("""<div class="field__input"><label><span ng-bind-html="(harrowFieldInput.label | translate | toTrusted)"/></label>#{transclude}<span/></div>""")

  @$element.replaceWith(el)
  @$element = @$compile(el)(@$scope.$new())

  @model = @$element.find('input,select,textarea').controller('ngModel')
  @input = @$element.find('input,select,textarea')
  @initWatches()

Controller::defaultType = () ->
  !/(checkbox|radio)/.test(@type)

Controller::initWatches = () ->
  @$scope.$watch 'harrowFieldInput.model.$modelValue', =>
    @clearServerErrors()
  @$scope.$watch 'harrowFieldInput.model.$valid', ($valid) =>
    if $valid
      @$element.removeClass('hasError')
    else if @model.$dirty
      @$element.addClass('hasError')

  @$scope.$watchCollection 'harrowFieldInput.model.$error', ($error) =>
    @errors = []
    angular.forEach $error, (isError, errorKey) =>
      if isError
        errorKey = errorKey.replace(/^server_/, '')
        error = @maybeTranslate("#{@translationRoot}.errors.#{@model.$name}.#{errorKey}")
        error ||= @$translate.instant("errors.field.#{errorKey}")
        @errors.push error
    @input.next('span').attr('data-error-messages', @errors)

Controller::maybeTranslate = (key) ->
  translation = @$translate.instant(key)
  return undefined if translation == key
  translation

Controller::clearServerErrors = () ->
  angular.forEach @model?.$error, (_, message) =>
    if message.match(/^server_/)
      @model.$setValidity message, true

Directive = () ->
  restrict: 'AE'
  scope: true
  transclude: true
  replace: true
  controller: 'harrowFieldInputCtrl'
  controllerAs: 'harrowFieldInput'
  template: '<div/>'


angular.module('harrowApp').controller 'harrowFieldInputCtrl', Controller
angular.module('harrowApp').directive 'harrowFieldInput', Directive
