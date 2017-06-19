describe('Config: serverErrorInterceptor', function () {
  var $httpBackend, $http, $rootScope
  beforeEach(angular.mock.inject(function (_$httpBackend_, _$http_, _$rootScope_) {
    $httpBackend = _$httpBackend_
    $http = _$http_
    $rootScope = _$rootScope_
  }))

  var serverErrorStatusCodes = [500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511, 599]
  var okStatusCodes = [100, 101, 102]
  var successStatusCodes = [200, 201, 202, 203, 204, 205, 206, 207, 208, 226]
  var redirectStatusCodes = [300, 302, 303, 304, 305, 307, 308]
  var clientErrorCodes = [400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 421, 422, 423, 426, 428, 429, 431, 451, 499]

  serverErrorStatusCodes.forEach(function (statusCode) {
    it('broadcasts "http.serverError" for "' + statusCode + '"', function (done) {
      $httpBackend.expectPOST('http://test.host/session').respond(statusCode, {'error': statusCode + ' Bad Gateway'})
      var hasEvent = false
      $rootScope.$on('http.serverError', function (event, status, data) {
        expect(status).toEqual(statusCode)
        expect(data).toEqual({'error': statusCode + ' Bad Gateway'})
        hasEvent = true
      })
      $http({
        url: 'http://test.host/session',
        method: 'POST'
      }).catch(function () {
        expect(hasEvent).toBeTruthy()
        done()
      })
      $httpBackend.flush()
    })
  })

  ;[].concat(okStatusCodes, clientErrorCodes).forEach(function (statusCode) {
    it('does not effect rejection "' + statusCode + '"', function (done) {
      $httpBackend.expectPOST('http://test.host/session').respond(statusCode, '')
      $http({
        url: 'http://test.host/session',
        method: 'POST'
      }).catch(function () {
        done()
      })
      $httpBackend.flush()
    })
  })

  successStatusCodes.forEach(function (statusCode) {
    it('does not effect resolve "' + statusCode + '"', function (done) {
      $httpBackend.expectPOST('http://test.host/session').respond(statusCode, '')
      $http({
        url: 'http://test.host/session',
        method: 'POST'
      }).then(function () {
        done()
      })
      $httpBackend.flush()
    })
  })
})
