describe('Resource: scheduledExecutionResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, scheduledExecutionResource) {
    $httpBackend = _$httpBackend_
    api = scheduledExecutionResource
  }))

  describe('.save', function () {
    it('POST to "/api/scheduled-executions"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/scheduled-executions/,
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
