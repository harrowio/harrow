angular.module('harrowApp').config ($stateProvider) ->
  $stateProvider
    .state 'billing',
      parent: 'organization.edit'
      url: '/billing'
      data:
        breadcrumbs: ['organization']
      views:
        "main@layout":
          controller: 'billingCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/billing/index.html'
      resolve:
        projects: (projects, organization) ->
          projects.filter (project) ->
            return project.subject.organizationUuid == organization.subject.uuid
        selectedPlan: (organization, billingPlanResource) ->
          planUuid = organization.subject.planUuid
          planUuid = null if angular.isString(planUuid) && planUuid.length == 0
          planUuid = 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3' unless planUuid
          billingPlanResource.find(planUuid)
        plans: (billingPlanResource) -> billingPlanResource.all()

    .state 'createCreditCard',
      parent: 'organization.edit'
      url: '/credit_cards?{planUuid}'
      views:
        "main@layout":
          controller: 'creditCardCtrl'
          controllerAs: 'ctrl'
          templateUrl: 'views/billing/edit.html'
      resolve:
        selectedPlan: ($stateParams, organization, billingPlanResource) ->
          planUuid = organization.subject.planUuid
          planUuid = $stateParams.planUuid if $stateParams.planUuid
          planUuid = null if angular.isString(planUuid) && planUuid.length == 0
          planUuid = 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3' unless planUuid
          billingPlanResource.find(planUuid)
