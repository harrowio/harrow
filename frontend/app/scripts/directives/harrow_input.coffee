app = angular.module("harrowApp")

Controller = (
  @$element
  @$scope
  @$translate
  $transclude
  @$attrs
) ->

  @hideLabel = false

  @translationRoot = @$scope.harrowForm?.translationRoot

  if @$attrs.hasOwnProperty('class')
    @additionalClasses = @$attrs.class

  if @$attrs.hasOwnProperty('noLabel')
    @hideLabel = true

  # transclude the embedded input
  # see also transclusion in harrow-form for some explanation
  @$scope.$evalAsync =>
    $transclude @$scope.$new(), (clone, scope) =>
      te = @$element.find(".transclude")
      te.append(clone)
      # find the transcluded input
      @maybeAddFormControl(te)
      @input = te.find("input,select,textarea")
      @input.after('<span>')
    @initWatches()

  @$scope.harrowForm?.addInput(@)
  @

# init the watches on the input's ngModel
Controller::initWatches = () ->
  @model = @input.controller("ngModel")
  formGroup = @$element.find(".form-group")

  placeholder = @maybeTranslate("#{@translationRoot}.placeholder.#{@model?.$name}")
  @input.attr("placeholder", placeholder)
  @label = @maybeTranslate("#{@translationRoot}.label.#{@model?.$name}")
  @hint  = @maybeTranslate("#{@translationRoot}.hint.#{@model?.$name}")

  # clear server error on changes
  @$scope.$watch "harrowInput.model.$modelValue", (v) =>
    @clearServerErrors()

  # Add has-error class depending on field validity
  @$scope.$watch "harrowInput.model.$valid", ($valid) =>
    if $valid
      @$element.find('.field__input').removeClass("hasError")
    else if @model and @model.$dirty
      @$element.find('.field__input').addClass("hasError")

  # Collect, strip server_ prefix and translate errors
  @$scope.$watchCollection "harrowInput.model.$error", ($error) =>
    @errors = []
    angular.forEach $error, (isError, errorKey) =>
      if isError
        # strip server_ prefix from the error key, if any
        errorKey = errorKey.replace(/^server_/, '')
        error = @maybeTranslate("#{@translationRoot}.errors.#{@model.$name}.#{errorKey}")
        error ||= @$translate.instant("errors.field.#{errorKey}")
        @errors.push error
    @input.parent().find('span').attr('data-error-messages', @errors)


# Clears all error keys prefixed with server_, to enable the user to re-submit the form.
Controller::clearServerErrors = () ->
  angular.forEach @model?.$error, (_, message) =>
    if message.match(/^server_/)
      @model.$setValidity message, true

# returns undefined if a translation key is not found, instead of the translation key itself
Controller::maybeTranslate = (key) ->
  t = @$translate.instant(key)
  if t == key
    undefined
  else
    t

#  Bootstrap requires to apply .form-control to most, but not all form elements:
#
# select
# textarea
# input[type="text"]
# input[type="password"]
# input[type="datetime"]
# input[type="datetime-local"]
# input[type="date"]
# input[type="month"]
# input[type="time"]
# input[type="week"]
# input[type="number"]
# input[type="email"]
# input[type="url"]
# input[type="search"]
# input[type="tel"]
# input[type="color"]
# see: https://github.com/twbs/bootstrap/blob/master/less/forms.less
#
# additionally, the class should be added to pre.ace_editor
Controller::maybeAddFormControl = (input) ->
  input.find("select,textarea").addClass("form-control")
  selectors = [
    "input:not([type])" # if no type is given, it defaults to "text"
    "input[type='text']"
    "input[type='password']"
    "input[type='datetime']"
    "input[type='datetime-local']"
    "input[type='date']"
    "input[type='month']"
    "input[type='time']"
    "input[type='week']"
    "input[type='number']"
    "input[type='email']"
    "input[type='url']"
    "input[type='search']"
    "input[type='tel']"
    "input[type='color']"
  ]
  input.find(selectors.join(",")).addClass("form-control")
  # wait for ui-ace to do its thing...
  @$scope.$evalAsync () ->
    input.find("pre.ace_editor").addClass("form-control")

HarrowInput = () ->
  restrict: "E"
  scope: true
  transclude: true
  replace: true
  controller: 'harrowInputCtrl'
  controllerAs: "harrowInput"
  templateUrl: 'views/directives/harrow_input.html'

app.controller('harrowInputCtrl', Controller)
app.directive("harrowInput", HarrowInput)
