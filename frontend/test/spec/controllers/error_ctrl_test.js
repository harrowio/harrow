describe('Controller: errorCtrl', function () {
  var ctrl, $httpBackend
  beforeEach(angular.mock.inject(function ($controller, $rootScope, _$httpBackend_, authentication) {
    $httpBackend = _$httpBackend_
    authentication.currentSession = {
      subject: {
        uuid: 'abc123'
      }
    }
    $rootScope.$intercom = {
      boot: function () {},
      update: function () {}
    }
    ctrl = $controller('errorCtrl')
    spyOn(ctrl.$state, 'go')
  }))

  describe('.resendVerificationEmail()', function () {
    it('performs HTTP POST to /verify-email', function (done) {
      $httpBackend.expect('POST', /\/api\/verify-email/, null, {
        'X-Harrow-Session-Uuid': 'abc123',
        'Accept': 'application/json, text/plain, */*'
      }).respond(201)
      ctrl.resendVerificationEmail().then(function () {
        expect(ctrl.$state.go).toHaveBeenCalledWith('errors/verification_email_sent')
        done()
      })
      $httpBackend.flush()
    })

    it('handles HTTP POST to /verify-email failure', function (done) {
      $httpBackend.expect('POST', /\/api\/verify-email/).respond(422)
      ctrl.resendVerificationEmail().then(function () {
        expect(ctrl.flash.error).toEqual('Failed to resend verification email.')
        done()
      })
      $httpBackend.flush()
    })
  })
})
