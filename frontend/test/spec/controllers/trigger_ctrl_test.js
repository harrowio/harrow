describe('Controller: triggerCtrl', function () {
  var ctrl, $scope, $state, $httpBackend, gitTriggerResource, $q
  beforeEach(angular.mock.inject(function (authentication, $controller, $rootScope, _$httpBackend_, _gitTriggerResource_, _$q_, _$state_) {
    $state = _$state_
    $q = _$q_
    $scope = $rootScope
    $httpBackend = _$httpBackend_
    gitTriggerResource = _gitTriggerResource_
    ctrl = $controller('triggerCtrl', {
      project: {
        subject: {
          uuid: 'abc123'
        }
      },
      triggerType: 'gitTrigger',
      trigger: {
        subject: {}
      },
      repositories: [],
      tasks: [
        {subject: {uuid: 'abc123'}}
      ],
      scripts: [],
      environments: [],
      task: null,
      triggerResource: gitTriggerResource
    })
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('.save()', function () {
    it('saves and redirects scope to "triggers"', function () {
      $httpBackend.expect('POST', /\/api\/git-triggers/).respond(200, {})
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('GET', /\/api\/organizations\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project_organization.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/git-triggers/).respond(200, jasmine.getJSONFixture('GET_api_project_git-triggers.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/schedules/).respond(200, jasmine.getJSONFixture('GET_api_project_schedules.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/webhooks/).respond(200, jasmine.getJSONFixture('GET_api_project_webhooks.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/environments/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/tasks/).respond(200, [])
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/jobs/).respond(200, [])
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/repositories/).respond(200, jasmine.getJSONFixture('GET_api_project_repositories.json'))

      ctrl.save()
      $httpBackend.flush()

      expect(ctrl.flash.success).toEqual('Saved Git Trigger')
      expect($state.current.name).toEqual('triggers')
    })
    it('saves and redirects scope to returnScope')
    it('rejects with error flash')
  })
})
