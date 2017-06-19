/*eslint-disable new-cap*/
describe('Resource: webhookResource', function () {
  var $httpBackend, api, model

  beforeEach(angular.mock.inject(function (_$httpBackend_, webhookResource) {
    $httpBackend = _$httpBackend_
    let data = {
      subject: {
        jobUuid: 'job123'
      }
    }
    api = webhookResource
    model = new api.model(data)
  }))

  describe('.model().subject', function () {
    it('redefines job as task', function () {
      model.subject.taskUuid = 'job123'
    })
  })

  describe('.save', function () {
    it('POST to "/api/webhooks"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/webhooks$/,
        jasmine.validateHttpParams({subject: {projectUuid: 'proj123', jobUuid: 'task123'}})
      ).respond(201)

      api.save({
        subject: {
          projectUuid: 'proj123',
          taskUuid: 'task123'
        }
      })
      $httpBackend.flush()
    })
  })
})
