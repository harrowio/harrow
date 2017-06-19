/*eslint-disable new-cap*/
describe('Resources: taskResource', function () {
  var api, model, data, $httpBackend
  beforeEach(angular.mock.inject(function (taskResource, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    data = jasmine.getJSONFixture('GET_api_job.json')
    api = taskResource
    model = new api.model(data)
  }))

  describe('.model() - linked resources', function () {
    it('defines "triggers()"', function () {
      $httpBackend.expect('GET', /\/api\/jobs\/[0-9a-f-]+\/triggers\/git/).respond(200, jasmine.getJSONFixture('GET_api_project_git-triggers.json'))
      $httpBackend.expect('GET', /\/api\/jobs\/[0-9a-f-]+\/triggers\/schedules/).respond(200, jasmine.getJSONFixture('GET_api_project_schedules.json'))
      $httpBackend.expect('GET', /\/api\/jobs\/[0-9a-f-]+\/triggers\/webhooks/).respond(200, jasmine.getJSONFixture('GET_api_project_webhooks.json'))
      $httpBackend.expect('GET', /\/api\/jobs\/[0-9a-f-]+\/triggers\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      expect(model.triggers).toBeDefined('has triggers')
      model.triggers().then(function (response) {
        expect(response.gitTriggers).toBeDefined()
        expect(response.schedules).toBeDefined()
        expect(response.webhooks).toBeDefined()
        expect(response.taskNotifiers).toBeDefined()
      })
      $httpBackend.flush()
    })

    it('defines "notifiers()"', function () {
      $httpBackend.expect('GET', /\/api\/tasks\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_task.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/slack-notifiers/).respond(200, jasmine.getJSONFixture('GET_api_project_slack-notifiers.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+\/job-notifiers/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/jobs\/[0-9a-f-]+\/notification-rules/).respond(200, jasmine.getJSONFixture('GET_api_job_notification-rules.json'))

      expect(model.notifiers).toBeDefined('has notifiers')
      model.notifiers().then(function (response) {
        expect(response.slackNotifiers).toBeDefined()
        expect(response.taskNotifiers).toBeDefined()
      })
      $httpBackend.flush()
    })
  })

  describe('.save', function () {
    it('POST to "/api/jobs"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/jobs/,
        jasmine.validateHttpParams({subject: {projectUuid: 'proj123', jobUuid: 'task123', taskUuid: 'script123'}})
      ).respond(201)

      api.save({
        subject: {
          projectUuid: 'proj123',
          taskUuid: 'task123',
          scriptUuid: 'script123'
        }
      })
      $httpBackend.flush()
    })
  })
})
