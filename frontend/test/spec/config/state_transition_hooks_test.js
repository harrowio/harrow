describe('Config: stateTransitionHooks', function () {
  var $state, $scope, $httpBackend
  beforeEach(angular.mock.inject(function (_$state_, _$httpBackend_, $rootScope, authentication) {
    $state = _$state_
    $scope = $rootScope
    $httpBackend = _$httpBackend_
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('returnState', function () {
    it('retains returnState params when given', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('GET', /\/api\/organizations\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project_organization.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/tasks/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/environments/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/slack-notifiers/).respond(200, jasmine.getJSONFixture('GET_api_project_slack-notifiers.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      $state.go('notifiers.slackNotifier', {projectUuid: 'abc123', returnState: 'dashboard'})
      $httpBackend.flush()
      expect($state.params.returnState).toEqual('dashboard')

      $state.go('notifiers', {projectUuid: 'abc123', returnTo: true})
      $scope.$digest()
      expect($state.current.name).toEqual('dashboard')
      expect($state.params.returnState).toEqual('')
    })
  })
})
