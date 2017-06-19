describe('Resource: scheduleResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, scheduleResource) {
    $httpBackend = _$httpBackend_
    api = scheduleResource
  }))

  describe('.save', function () {
    it('POST to "/api/schedule"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/schedule/,
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
