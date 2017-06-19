describe('Controller: loginCtrl', function () {
  var ctrl, $scope, $state, $httpBackend, $stateParams
  beforeEach(angular.mock.inject(function ($controller, $rootScope, _$state_, _$httpBackend_, _$stateParams_) {
    $state = _$state_
    $scope = $rootScope
    $stateParams = _$stateParams_
    $httpBackend = _$httpBackend_
    ctrl = $controller('loginCtrl', {
      $scope: $scope
    })
    spyOn($state, 'go')
    spyOn(ctrl.authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('login route', function () {
    it('login and redirects to "dashboard"', function () {
      ctrl.user = {
        subject: {
          email: 'test@localhost',
          password: 'changeme123'
        }
      }
      $httpBackend.expect('POST', /\/api\/sessions/).respond(200, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /\/api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.login()
      $httpBackend.flush()

      expect($state.go).toHaveBeenCalledWith('dashboard')
    })

    it('fails to authenticate', function () {
      ctrl.user = {
        subject: {
          email: 'test@localhost',
          password: 'changeme123'
        }
      }
      $httpBackend.expect('POST', /\/api\/sessions/).respond(422, jasmine.getJSONFixture('PUT_api_sessions.json'))

      ctrl.login()
      $httpBackend.flush()

      expect(ctrl.flash.error).toEqual('Email address or password was not valid')
    })
  })

  describe('Two Factor Authentication', function () {
    it('redirects to "session_confirmation"', function () {
      ctrl.user = {
        subject: {
          email: 'test@localhost',
          password: 'changeme123'
        }
      }
      ctrl.authentication.hasValidSession.and.callFake(function () {
        return false
      })
      $httpBackend.expect('POST', /\/api\/sessions/).respond(200, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /\/api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.login()
      $httpBackend.flush()

      expect($state.go).toHaveBeenCalledWith('session_confirmation', jasmine.any(Object))
    })
  })

  describe('invitation', function () {
    it('login and recirects to "invitations/show"', function () {
      ctrl.user = {
        subject: {
          email: 'test@localhost',
          password: 'changeme123'
        }
      }
      ctrl.invitationUuid = 'abc123'

      $httpBackend.expect('POST', /\/api\/sessions/).respond(200, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /\/api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.login()
      $httpBackend.flush()

      expect($state.go).toHaveBeenCalledWith('invitations/show', {uuid: 'abc123'})
    })
  })

  describe('guest access', function () {
    it('creates a guest account and directs to wizard', function () {
      $httpBackend.expect('POST', /api\/users/, jasmine.validateHttpParams({subject: {signupParameters: {isGuest: true}}})).respond(201, jasmine.getJSONFixture('GET_api_user.json'))
      $httpBackend.expect('POST', /api\/sessions/).respond(201, jasmine.getJSONFixture('PUT_api_sessions.json'))
      $httpBackend.expect('GET', /api\/users/).respond(200, jasmine.getJSONFixture('GET_api_user.json'))

      ctrl.guest()
      $httpBackend.flush()

      expect($state.go).toHaveBeenCalledWith('dashboard')
    })
  })
})
