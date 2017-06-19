braintree = require('braintree-web')

Controller = (
  @organization
  @selectedPlan
  @$http
  @$q
  @$state
  @flash
  @$translate
  @endpoint
  Stateful
) ->
  @stateful = new Stateful()

  @statefulOptions = {}

  @stateful.on 'purchasing', =>
    @statefulOptions.attrs =
      class: 'btn'
      ngDisabled: true
    @statefulOptions.content = '<span svg-icon="icon-spinner"></span> Processing'

  @stateful.on 'paid', =>
    @statefulOptions.attrs =
      class: 'btn btn--green'
      ngDisabled: true
    @statefulOptions.content = '<span svg-icon="icon-complete-alt"></span> Fantastic!'

  @stateful.on 'error', =>
    @statefulOptions.attrs =
      class: 'btn btn--primary'
      ngDisabled: false
    @statefulOptions.content = '<span svg-icon="icon-error-alt"></span> Try again?'

  @braintree = {}

  @

Controller::saveCard = () ->
  @stateful.transitionTo('purchasing')
  @$http.post("#{@endpoint}/billing-plans/braintree/client-token")
  .then (res) =>
    braintreeClient = new braintree.api.Client({
      clientToken: res.data.clientToken
    })
    deferred = @$q.defer()
    console.log @braintree
    braintreeClient.tokenizeCard @braintree, (err, nonce) ->
      if err
        deferred.reject(err)
      else
        deferred.resolve(nonce)
    deferred.promise
  .then (nonce) =>
    @organization.addCreditCard(nonce, @organization.subject.uuid)
  .then () =>
    @purchase()
  .then =>
    @stateful.transitionTo('paid')
    @$state.go('billing', {uuid: @organization.subject.uuid, organizationUuid: @organization.subject.uuid}, {reload: true})
    return
  .catch (reason) =>
    @stateful.transitionTo('error')
    @flash.error = reason.message
    @$q.reject(reason)

Controller::purchase = () ->
  @selectedPlan.purchase(@organization.subject.uuid).then =>
    @flash.success = @$translate.instant('forms.billing.flashes.success')


angular.module('harrowApp').controller 'creditCardCtrl', Controller

