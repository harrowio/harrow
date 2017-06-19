describe('Resource: gitTriggerResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, gitTriggerResource) {
    $httpBackend = _$httpBackend_
    api = gitTriggerResource
  }))

  describe('.save', function () {
    it('POST to "/api/git-triggers"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/git-triggers/,
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
