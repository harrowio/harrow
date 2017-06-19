/*eslint-disable new-cap*/
describe('Resources: projectResource', function () {
  var api, model, data, $httpBackend
  beforeEach(angular.mock.inject(function (projectResource, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    data = jasmine.getJSONFixture('GET_api_project.json')
    api = projectResource
    model = new api.model(data)
  }))
  describe('.model() - linked resources', function () {
    it('defines "notifiers()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/slack-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      expect(model.notifiers).toBeDefined('has notifiers')
      model.notifiers().then(function (response) {
        expect(response.slackNotifiers).toBeDefined()
        expect(response.taskNotifiers).toBeDefined()
      })
      $httpBackend.flush()
    })

    it('defines "triggers()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/git-triggers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/schedules/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/webhooks/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      expect(model.triggers).toBeDefined('has triggers')
      model.triggers().then(function (response) {
        expect(response.gitTriggers).toBeDefined()
        expect(response.webhooks).toBeDefined()
        expect(response.schedules).toBeDefined()
        expect()
      })
      $httpBackend.flush()
    })

    it('defines "slackNotifiers()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/slack-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.slackNotifiers).toBeDefined('has slackNotifiers')
      model.slackNotifiers()
      $httpBackend.flush()
    })

    it('defines "environments()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/environments/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.environments).toBeDefined('has environments')
      model.environments()
      $httpBackend.flush()
    })

    it('defines "tasks()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.environments).toBeDefined('has tasks')
      model.tasks()
      $httpBackend.flush()
    })

    it('defines "tasks()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.tasks).toBeDefined('has tasks')
      model.tasks()
      $httpBackend.flush()
    })

    it('defines "repositories()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/repositories/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.repositories).toBeDefined('has repositories')
      model.repositories()
      $httpBackend.flush()
    })

    it('defines "organization()"', function () {
      $httpBackend.expect('GET', /\/api\/organizations\/[0-9a-f-]+/).respond(200, {})
      expect(model.environments).toBeDefined('has organization')
      model.organization()
      $httpBackend.flush()
    })

    it('defines "scripts()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/tasks/).respond(200, {
        subject: {
          body: ''
        }
      })
      expect(model.scripts).toBeDefined('has scripts')
      model.scripts()
      $httpBackend.flush()
    })

    it('defines "gitTriggers()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/git-triggers/).respond(200, {})
      expect(model.gitTriggers).toBeDefined('has gitTriggers')
      model.gitTriggers()
      $httpBackend.flush()
    })

    it('defines "webhooks()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/webhooks/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.webhooks).toBeDefined('has webhooks')
      model.webhooks()
      $httpBackend.flush()
    })

    it('defines "members()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/members/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.members).toBeDefined('has members')
      model.members()
      $httpBackend.flush()
    })

    it('defines "operations()"', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/operations/).respond(200, {
        subject: {}
      })
      expect(model.operations).toBeDefined('has operations')
      model.operations()
      $httpBackend.flush()
    })

    it('defines "leave()"', function () {
      $httpBackend.expect('DELETE', /\/api\/projects\/[0-9a-f-]+\/members/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      expect(model.leave).toBeDefined('has leave')
      model.leave()
      $httpBackend.flush()
    })
  })
})
