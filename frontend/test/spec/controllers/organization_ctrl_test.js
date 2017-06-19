describe('Controller: creditCardCtrl', function () {
  var ctrl, $scope, $state, $httpBackend, $q, org, plan
  beforeEach(angular.mock.inject(function (authentication, $controller, $rootScope, _$httpBackend_, _$q_, _$state_, Organization, BillingPlan) {
    $state = _$state_
    $q = _$q_
    $scope = $rootScope
    $httpBackend = _$httpBackend_
    org = new Organization(jasmine.getJSONFixture('GET_api_project_organization.json'))
    plan = new BillingPlan(jasmine.getJSONFixture('GET_api_billing_plan.json'))
    ctrl = $controller('creditCardCtrl', {
      organization: org,
      plans: null,
      selectedPlan: plan
    })
    ctrl.braintree = {
      number: "4556969783321900",
      cvv: "082",
      expirationMonth: 02,
      expirationYear: 2020,
      cardHolderName: "Max Mustermann"
    }
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('.saveCard', function () {
    fit('creates a nonce, tokenizesCard and adds card to organization', function () {
      $httpBackend.when('GET', /api\/organizations\/[0-9a-f-]+\/projects/).respond(200, jasmine.getJSONFixture('GET_api_user_projects.json'))
      $httpBackend.when('GET', /api\/projects\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('GET_api_project_jobs.json'))
      $httpBackend.when('GET', /api\/projects\/[0-9a-f-]+\/operations/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.when('GET', /api\/projects\/[0-9a-f-]+\/repositories/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.when('GET', /api\/environments\/[0-9a-f-]+/).respond(200, {})
      $httpBackend.when('GET', /api\/tasks\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_task.json'))
      $httpBackend.when('GET', /api\/organizations\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project_organization.json'))

      $httpBackend.expect('POST', /\/api\/billing-plans\/braintree\/client-token/).respond(201, jasmine.getJSONFixture('POST_api_billing_plans_braintree_client_token.json'))
      $httpBackend.expect('POST', /api\/billing-plans\/braintree\/credit-cards/).respond(201, {})
      $httpBackend.expect('POST', /api\/billing-plans\/braintree\/purchase/).respond(201)

      ctrl.saveCard()
      $httpBackend.flush()
      expect(ctrl.flash.success).toEqual('Billing plan changed successfully')
    })
  })
})

