Controller = (
  @ga
  @stencilResource
  @project
  @$state
  @flash
  @$translate
  @$q
  Stateful
  @$scope
) ->
  @stateful = new Stateful()
  @stateful.on 'pending', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-spinner"></span> Saving'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'complete', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-spinner"></span> Saving'
      attrs:
        class: 'btn'
        ngDisabled: true
  @stateful.on 'error', =>
    @saveButtonOptions =
      content: '<span svg-icon="icon-error-alt"></span> Error'
      attrs:
        class: "btn btn--primary"
        ngDisabled: false
  @

Controller::applyStencil = (id) ->
  return if @stateful.state.pending
  @stateful.transitionTo('pending')
  @ga 'send', 'event', 'wizard', 'stencils', 'prefillSubmitted'
  stencil =
    subject:
      projectUuid: @project.subject.uuid
      id: id

  @stencilResource.save(stencil).then =>
    @$scope.$parent.ctrl?.progress = 2 # mark track as complete
    @stateful.transitionTo('complete', {terminal: true})
    @ga 'send', 'event', 'wizard', 'stencils', 'prefillSuccess'
    @flash.success = @$translate.instant('forms.wizard.stencils.flashes.success')
    @$state.go('wizard.project.finished', {projectUuid: @project.subject.uuid})
    return
  .catch (reason) =>
    @stateful.transitionTo('error', {terminal: true})
    @ga 'send', 'event', 'wizard', 'stencils', 'prefillError'
    @flash.error = @$translate.instant('forms.wizard.stencils.flashes.fail')
    @$q.reject(reason)

angular.module('harrowApp').controller 'wizardStencilsCtrl', Controller
