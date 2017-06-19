describe('Resource: featureResource', function () {
  var $httpBackend, api

  beforeEach(angular.mock.inject(function (_$httpBackend_, appFeatureResource) {
    $httpBackend = _$httpBackend_
    api = appFeatureResource
  }))

  describe('.all', function () {
    it('GET to "/api/api-features"', function () {
      $httpBackend.expect(
        'GET',
        /\/api\/api-features/
      ).respond(200, {
        collection: []
      })

      api.all()
      $httpBackend.flush()
    })
  })
})
