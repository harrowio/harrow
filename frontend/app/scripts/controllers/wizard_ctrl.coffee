app = angular.module("harrowApp")

WizardCtrl = (
  @organizationResource
  @projectResource
  @$translate
  @flash
  @$state
  @menuItems
  $scope
) ->
  @menu = @menuItems.wizard
  @progress = 0
  @stations = @$state.current.data?.sectionsToComplete || 0
  $scope.$watch =>
    @$state.current.name
  , (stateName) =>
    if @$state.current.data && @$state.current.data.completedSections
      progress = @$state.current.data.completedSections.length
      if @$state.includes('wizard.project.connect.repo')
        @progress = (progress / (@stations - 1.5))
      else
        @progress = (progress / (@stations - 1))
  @

WizardCtrl::connect = (provider) ->


WizardCtrl::iconColor = (sref) ->
  return @$state.is(sref) ||  @$state.current.data.completedSections?.indexOf(sref) >= 0

app.controller("wizardCtrl", WizardCtrl)
