app = angular.module("harrowApp")

Controller = (
  $element
  $scope
  $attrs
  $transclude
) ->

  @translationRoot = $attrs.translationRoot

  # Here be dragons
  #
  # Transclude manually to have a chance to specify the transclusion scope.
  #
  # Although harrow-input, harrow-select are a child DOM element, because of transclusion they don't
  # create child scopes of harrow-form by default, but child scopes of harrow-forms parent (IOW: they are siblings).
  # Passing $scope as the first parameter changes this.
  #
  # see: https://github.com/angular/angular.js/commit/683fd713c41eaf5da8bfbf53b574e0176c18c518#diff-5f00714557b95b18ae74853f8a027b56R53
  # "
  # use evalAsync so that we don't process transclusion before directives on the parent element even when the
  # transclusion replaces the current element. (we can't use priority here because that applies only to compile fns
  # and not controllers
  # "
  #
  # Without evalAsync there was a race condition that made the transcluded element's controllers to be invoked before
  # this directive is finished, leading to missing/broken state on harrowInput.form, for example.
  # debugger
  $scope.$evalAsync ->
    $transclude $scope.$new(), (clone, scope) ->

      $element.find(".transclude").replaceWith(clone)

      if $attrs.hasOwnProperty('noControls')
        $element.find('.form-actions').remove()
      else
        secondaryButtons = $element.find('.btn:not(.btn--primary)').detach()
        secondaryButtons.appendTo($element.find(".form-actions"))

  # get reference to save function
  @save = () ->
    $scope.$eval($attrs.save)

  # HarrowInputs will register here
  @inputs = []
  @

Controller::addInput = (input) ->
  @inputs ||= []
  @inputs.push input

Controller::clearAllServerErrors = ->
  angular.forEach @inputs, (input) ->
    input.clearServerErrors()

Controller::submit = ->
  res = @save()
  if res?.catch
    # ...assuming we got a promise
    res.catch (object) =>
      @applyErrors(object.errors) if object?.errors
    res.then () =>
      @form?.$setPristine()
  else
    res

Controller::applyErrors = (errors) ->
  angular.forEach errors, (messages, fieldName) =>
    angular.forEach messages, (message) =>
      if @form[fieldName]
        @form[fieldName].$setValidity("server_#{message}", false)

HarrowForm = () ->
  restrict: "E"
  scope: true
  transclude: true
  replace: true
  controller: 'harrowFormCtrl'
  controllerAs: "harrowForm"
  templateUrl: 'views/directives/harrow_form.html'

app.controller('harrowFormCtrl', Controller)
app.directive("harrowForm", HarrowForm)
