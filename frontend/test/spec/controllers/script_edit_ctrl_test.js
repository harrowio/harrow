describe('Controller: ScriptEditCtrl', function () {
  var $ctrl, $scope, $httpBackend, testScript, $state
  beforeEach(angular.mock.inject(function ($controller, $rootScope, _$state_, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    $state = _$state_
    $scope = $rootScope.$new()

    testScript = {
      script: {
        body: '#!/bin/bash -e\ndate'
      }
    }
    var script = {
      subject: {
        uuid: 'abc123'
      }
    }
    var project = {
      subject: {
        uuid: 'abc123'
      }
    }
    var environments = []
    var repositories = []
    var secrets = []
    var environment = {
      subject: {
        uuid: 'abc1234',
        name: 'Default',
        variables: {}
      }
    }
    environments.push(environment)

    testScript.environment = environment.subject
    $ctrl = $controller('scriptEditCtrl', {
      $scope: $scope,
      script: script,
      testScript: testScript,
      project: project,
      environments: environments,
      repositories: repositories,
      secrets: secrets
    })
    spyOn($state, 'go')
  }))

  describe('.save()', function () {
    it('saves a task (script + environment)', function () {
      $httpBackend.expect('POST', /api\/script-editor\/save/).respond(204, '')
      $httpBackend.expect('POST', /api\/jobs/, jasmine.validateHttpParams({subject: {uuid: jasmine.any(String), scriptUuid: 'abc123', environmentUuid: 'abc1234', description: 'autogenerated by script', name: 'autogenerated script'}})).respond(200, jasmine.getJSONFixture('GET_api_job.json'))
      $ctrl.save()
      $httpBackend.flush()
      expect($state.go).toHaveBeenCalledWith('script', {projectUuid: 'abc123', scriptUuid: 'abc123'}, {reload: true})
    })
  })
})
