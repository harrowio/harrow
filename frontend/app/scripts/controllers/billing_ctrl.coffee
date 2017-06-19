Controller = (
  @organization
  @projects
  @plans
  @limits
  @$state
  @$translate
  @flash
  @$filter
  Stateful
) ->
  @statefulOptions = {
    content: ''
  }
  @stateful = new Stateful()
  @stateful.on 'purchasing', =>
    @statefulOptions.content = '<span svg-icon="icon-spinner"></span> Processing'
  @

Controller::isCurrentPlan = (plan) ->
  @organization.subject.planUuid == plan.subject.uuid

Controller::isPaid = () ->
  @organization.subject.planUuid != 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3'

Controller::currentPlan = () ->
  @plans.find (plan) =>
    @organization.subject.planUuid == plan.subject.uuid

Controller::isPlanUpgrade = (plan) ->
  return true if !@currentPlan()
  plan.subject.pricePerMonth > @currentPlan().subject.pricePerMonth

Controller::isRecommendedPlan = (plan) ->
  # plan.subject.usersIncluded > @organization.members.
   @organization._embedded.limits[0].subject.exceeded && plan.subject.projectsIncluded >= @projects.length

Controller::selectPlan = (plan) ->
  return if plan.subject.uuid == @organization.subject.planUuid
  return if @pendingPlanUuid
  if @organization.subject.creditCards.length > 0
    pricePerMonth = @$filter('currency')(plan.subject.pricePerMonth, '$', 0)
    if plan.subject.pricePerMonth == 0
      pricePerMonth = 'Free'
    upgradeMessage = @$translate.instant('forms.billing.prompt.upgrade', {pricePerMonth: pricePerMonth})
    downgradeMessage = @$translate.instant('forms.billing.prompt.downgrade', {pricePerMonth: pricePerMonth})
    message = if @isPlanUpgrade(plan) then upgradeMessage else downgradeMessage
    if confirm(message)
      @purchase(plan)
  else
    @$state.go('createCreditCard', {planUuid: plan.subject.uuid})
  return

Controller::purchase = (plan) ->
  @stateful.transitionTo('purchasing')
  @pendingPlanUuid = plan.subject.uuid
  plan.purchase(@organization.subject.uuid).then =>
    @pendingPlanUuid = null
    @organization.subject.planUuid = plan.subject.uuid
    @stateful.transitionTo('paid')
    @flash.success = @$translate.instant('forms.billing.flashes.success')
    @$state.go('billing', {uuid: @organization.subject.uuid}, {reload: true})
    return
  .catch (reason) =>
    @pendingPlanUuid = null
    @stateful.transitionTo('error')
    @flash.error = @$translate.instant('forms.billing.flashes.fail')
    @$q.reject(reason)

Controller::freePlan = () ->
  @plans.find (plan) ->
    plan.subject.uuid == 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3'

Controller::hasTrialEnded = () ->
  !@isPaid() && @limits.subject.exceeded && @limits.subject.trialDaysLeft <= 0

Controller::isTrialEnding = () ->
  !@isPaid() && @limits.subject.exceeded && @limits.subject.trialDaysLeft <= 7 && @limits.subject.trialDaysLeft > 0

angular.module('harrowApp').controller 'billingCtrl', Controller
