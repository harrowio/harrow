describe('Resource: taskNotifierResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, taskNotifierResource) {
    $httpBackend = _$httpBackend_
    api = taskNotifierResource
  }))

  describe('.save', function () {
    it('POST to "/api/job-notifiers"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/job-notifiers/,
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
