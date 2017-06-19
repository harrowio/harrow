Controller = (
  @projects
  @organization
  @cards
  @flash
  @$state
  @$translate
  @organizationResource
) ->
  @

Controller::save = () ->
  @organizationResource.save(@organization).then () =>
    @flash.success = @$translate.instant('forms.organizationForm.flashes.success')
    @$state.reload()
    return

Controller::delete = () ->
  @organizationResource.delete(@organization.subject.uuid).then () =>
    @flash.success = @$translate.instant('forms.organizationForm.flashes.delete.success')
    @$state.go('dashboard', {}, {reload: true})
    return
  .catch () =>
    @flash.error = @$translate.instant('forms.organizationForm.flashes.delete.fail')
    return

angular.module("harrowApp").controller "organizationShowCtrl", Controller
