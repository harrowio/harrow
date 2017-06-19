app = angular.module("harrowApp")

app.factory "BillingPlan", ($injector, $http) ->
  BillingPlan = (data) ->
    $.extend(true, @, data)
    @subject.pricePerMonth = parseFloat @subject.pricePerMonth
    @subject.pricePerAdditionalUser = parseFloat @subject.pricePerAdditionalUser
    @$http = $http
    @

  BillingPlan::purchase = (organizationUuid) ->
    @$http.post(@_links['braintree-purchase'].href, {
      organizationUuid: organizationUuid,
      planUuid: @subject.uuid,
    })

  BillingPlan

app.factory "billingPlanResource", (Resource, BillingPlan) ->
  BillingPlanResource = () ->
    Resource.call(@)
    @

  BillingPlanResource:: = Object.create(Resource::)
  BillingPlanResource::basepath = "/billing-plans"
  BillingPlanResource::model = BillingPlan

  new BillingPlanResource()
