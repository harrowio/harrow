describe('Controller: signupCtrl', function () {
  var ctrl, $scope, $httpBackend
  beforeEach(angular.mock.inject(function (authentication, $controller, $rootScope, _$httpBackend_) {
    $httpBackend = _$httpBackend_
    $scope = $rootScope
    ctrl = $controller('signupCtrl', {})

    spyOn(ctrl.$state, 'go')
    spyOn(ctrl, 'ga')
    ctrl.user = {
      subject: {}
    }
  }))
  describe('.signup()', function () {
    it('creates account and redirects to dashboard', function () {
      $httpBackend.expect('POST', /api\/users/).respond(201, jasmine.getJSONFixture('GET_api_user.json'))
      $httpBackend.expect('POST', /api\/sessions/).respond(201, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.signup()
      $httpBackend.flush()

      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'user', 'signup-ok')
      expect(ctrl.stateful.state).toEqual('success')
      expect(ctrl.$state.go).toHaveBeenCalledWith('dashboard')
    })

    it('redirects to blocked', function () {
      $httpBackend.expect('POST', /api\/users/).respond(201, jasmine.getJSONFixture('GET_api_user.json'))
      $httpBackend.expect('POST', /api\/sessions/).respond(201, {subject: {blocks: ['validate-email']}})

      ctrl.signup()
      $httpBackend.flush()

      expect(ctrl.$state.go).toHaveBeenCalledWith('errors/blocked')
    })

    it('redirects to invitations', function () {
      ctrl.invitationUuid = 'abc123'
      $httpBackend.expect('POST', /api\/users/).respond(201, jasmine.getJSONFixture('GET_api_user.json'))
      $httpBackend.expect('POST', /api\/sessions/).respond(201, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.signup()
      $httpBackend.flush()

      expect(ctrl.$state.go).toHaveBeenCalledWith('invitations/show', {uuid: 'abc123'})
    })

    it('fails and sets state to "error"', function () {
      $httpBackend.expect('POST', /api\/users/).respond(422, jasmine.getJSONFixture('GET_api_user.json'))
      ctrl.signup()
      $httpBackend.flush()
      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'user', 'signup-failed')
      expect(ctrl.stateful.state).toEqual('error')
    })
  })

  describe('.githubSignin', function () {
    it('calls to oauth ', function () {
      spyOn(ctrl.oauth, 'signinGithub')
      ctrl.githubSignin()
      expect(ctrl.oauth.signinGithub).toHaveBeenCalled()
    })
  })
})
