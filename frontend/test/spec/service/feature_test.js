describe('Service: feature', function () {
  var api, $httpBackend
  beforeEach(angular.mock.inject(function (feature, _$httpBackend_) {
    api = feature
    $httpBackend = _$httpBackend_
  }))

  describe('loadFeatures', function () {
    it('performs a XHR request to the server', function () {
      $httpBackend.expect('GET', /\/api\/api-features/).respond(200, jasmine.getJSONFixture('GET_api_app_features.json'))
      api.loadFeatures()
      $httpBackend.flush()
    })
  })

  describe('isEnabled', function () {
    it('checks if feature is enabled', function () {
      api.enabledFeatures = [
        {
          subject: {
            name: 'oauth',
            enableAt: '1970-01-01T00:00:00.000Z'
          }
        }
      ]
      expect(api.isEnabled('oauth')).toBeTruthy()
    })

    xit('enables hierarchy', function () {
      api.enabledFeatures = [
        {
          subject: {
            name: 'oauth.github',
            enableAt: '1970-01-01T00:00:00.000Z'
          }
        }
      ]
      expect(api.isEnabled('oauth.*')).toBeTruthy()
    })

    it('features are disbaled by default', function () {
      expect(api.isEnabled('oauth')).toBeFalsy()
    })
  })

  describe('isDisabled', function () {
    it('features are disbaled by default', function () {
      expect(api.isDisabled('oauth')).toBeTruthy()
    })

    it('checks if feature is enabled', function () {
      api.enabledFeatures = jasmine.getJSONFixture('GET_api_app_features.json').collection
      expect(api.isDisabled('oauth')).toBeFalsy()
    })

    it('features are disbaled if enableAt is in the future', function () {
      api.enabledFeatures = [
        {
          subject: {
            name: 'oauth',
            enableAt: '2970-01-01T00:00:00.000Z'
          }
        }
      ]
      expect(api.isDisabled('oauth')).toBeTruthy()
    })
  })
})
