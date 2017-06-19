describe('Controller: wizardStencilsCtrl', function () {
  var $scope, $state, authentication, $httpBackend, $stateParams, ctrl

  beforeEach(angular.mock.inject(function ($rootScope, $controller, _$state_, _$stateParams_, _authentication_, _$httpBackend_) {
    $scope = $rootScope.$new()
    $state = _$state_
    $stateParams = _$stateParams_
    $stateParams.projectUuid = 'abc123'
    authentication = _authentication_
    $httpBackend = _$httpBackend_
    ctrl = $controller('wizardStencilsCtrl', {
      $scope: $scope,
      project: {
        subject: {
          uuid: 'abc123'
        }
      }
    })
    spyOn(ctrl, 'ga')

    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('routing', function () {
    it('redirects to "login" on unauthenticated users', function () {
      authentication.hasValidSession.and.callFake(function () {
        return false
      })
      $state.go('wizard.project.stencils', {projectUuid: 'abc123'})
      $scope.$digest()
      expect($state.current.name).toEqual('login')
    })

    it('routes wizardCtrl for authenticated users', function () {
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200)

      $state.go('wizard.project.stencils', {projectUuid: 'abc123'})
      $httpBackend.flush()

      expect($state.current.name).toEqual('wizard.project.stencils')
      expect($state.current.data.showViews).toEqual([])
      expect($state.current.views['main@wizard'].templateUrl).toEqual('views/wizard/stencils.html')
      expect($state.current.views['main@wizard'].controller).toEqual('wizardStencilsCtrl')
      expect($state.current.views['main@wizard'].controllerAs).toEqual('ctrl')
    })
  })

  describe('.applyStencil()', function () {
    it('successfully saves script', function (done) {
      $httpBackend.expect('POST', /\/api\/stencils/, jasmine.validateHttpParams({subject: {id: 'capistrano-rails'}})).respond(200)
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json')) // from the redirect

      ctrl.applyStencil('capistrano-rails').then(function () {
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'stencils', 'prefillSubmitted')
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'stencils', 'prefillSuccess')
        // expect($state.current.name).toEqual('wizard.project.users')
        expect($stateParams.projectUuid).toEqual('abc123')
        expect(ctrl.flash.success).toEqual('Stencil saved')
        done()
      })
      $httpBackend.flush()
    })

    it('fails to save script', function (done) {
      $httpBackend.expect('POST', /\/api\/stencils/).respond(422)
      ctrl.applyStencil('capistrano-rails').catch(function () {
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'stencils', 'prefillSubmitted')
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'stencils', 'prefillError')
        expect(ctrl.flash.error).toEqual('Unable to save Stencil')
        done()
      })
      $httpBackend.flush()
    })
  })
})
