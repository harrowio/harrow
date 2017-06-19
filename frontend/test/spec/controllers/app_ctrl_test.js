describe('Controller: appCtrl', function () {
  var ctrl, $httpBackend
  beforeEach(angular.mock.inject(function ($controller, $rootScope, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    $rootScope.$intercom = {
      boot: function () {},
      update: function () {}
    }
    ctrl = $controller('appCtrl')
    spyOn(ctrl.$state, 'go')
  }))

  it('clears all errors on success', function () {
    ctrl.flash.error = 'bad'
    ctrl.flash.success = 'good'
    expect(ctrl.flash.error).toBeNull()
    expect(ctrl.flash.success).toEqual('good')
  })

  describe('http.serverError', function () {
    var $http
    beforeEach(angular.mock.inject(function (_$http_) {
      $http = _$http_
    }))
    it('displays flash for permission errors', function () {
      $httpBackend.expectGET(/\/api\/auth-required/).respond(403, {
        reason: 'permission_denied'
      })
      $http.get('https://test.host/api/auth-required')
      $httpBackend.flush()
      expect(ctrl.flash.info).toEqual('You do not have permission to access this page.')
    })
  })

  describe('.logout()', function () {
    it('calls authentication.logout and redirects', function () {
      spyOn(ctrl.authentication, 'logout')
      ctrl.logout()
      expect(ctrl.authentication.logout).toHaveBeenCalled()
      expect(ctrl.$state.go).toHaveBeenCalledWith('login')
    })
  })
})
