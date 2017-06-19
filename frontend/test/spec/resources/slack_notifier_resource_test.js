describe('Resource: slackNotifierResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, slackNotifierResource) {
    $httpBackend = _$httpBackend_
    api = slackNotifierResource
  }))

  describe('.save', function () {
    it('POST to "/api/slack-notifiers"', function () {
      $httpBackend.expect(
        'POST',
        /\/api\/slack-notifiers/,
        jasmine.validateHttpParams({subject: {projectUuid: 'proj123'}})
      ).respond(201)

      api.save({
        subject: {
          projectUuid: 'proj123'
        }
      })
      $httpBackend.flush()
    })
  })
})
