describe('Controller: notifierCtrl', function () {
  var ctrl, $injector, $httpBackend, authentication, $state
  beforeEach(angular.mock.inject(function ($controller, _$injector_, _$httpBackend_, _authentication_, _$state_, slackNotifierResource) {
    $httpBackend = _$httpBackend_
    $state = _$state_
    $injector = _$injector_
    authentication = _authentication_
    ctrl = $controller('notifierCtrl', {
      project: {
        subject: {
          uuid: 'abc123'
        }
      },
      notifier: {
        subject: {
          projectUuid: 'abc123'
        }
      },
      tasks: null,
      scripts: null,
      environments: null,
      notifierType: 'slackNotifier',
      notifierResource: slackNotifierResource
    })
    spyOn(ctrl, 'ga')
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('.save()', function () {
    it('saves "slack" notification', function () {
      $httpBackend.expect('POST', /\/api\/slack-notifier/).respond(200, {})
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('GET', /\/api\/organizations\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project_organization.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/tasks/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/environments/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/slack-notifiers/).respond(200, jasmine.getJSONFixture('GET_api_project_git-triggers.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      var resource = $injector.get('slackNotifierResource')
      spyOn(resource, 'save').and.callThrough()

      ctrl.notifier.subject.notifierType = 'slack'
      ctrl.notifier.subject.name = '#example'
      ctrl.notifier.subject.webhookURL = 'http://slack.test.host/#example'

      ctrl.save()
      $httpBackend.flush()

      expect($state.current.name).toEqual('notifiers')
      expect(resource.save).toHaveBeenCalledWith({
        subject: jasmine.objectContaining({
          projectUuid: 'abc123',
          name: '#example',
          webhookURL: 'http://slack.test.host/#example'
        })
      })
    })
  })
})
