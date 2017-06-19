describe('Controller: wizardQuickStartCtrl', function () {
  var ctrl, $httpBackend, $scope
  beforeEach(angular.mock.inject(function ($controller, _$httpBackend_, $rootScope) {
    $scope = $rootScope
    ctrl = $controller('wizardQuickStartCtrl', {$scope: $scope})
    $httpBackend = _$httpBackend_
    spyOn(ctrl, 'ga')
    spyOn(ctrl.$state, 'go')
  }))

  describe('.save', function () {
    it('creates organization, project, repository and redirects to stencils', function () {
      $httpBackend.expect('POST', /\/api\/organizations/, jasmine.validateHttpParams({subject: {name: 'Harrow', public: false, planUuid: 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3'}})).respond(201, jasmine.getJSONFixture('GET_api_organization.json'))
      $httpBackend.expect('POST', /\/api\/projects/, jasmine.validateHttpParams({subject: {name: 'Frontend', organizationUuid: 'f3fcd172-ea46-4573-8195-3e211fa897c3'}})).respond(201, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('POST', /\/api\/repositories/, jasmine.validateHttpParams({subject: {url: 'https://github.com/harrow/frontend.git', projectUuid: '42e80914-ddad-4bb7-3ee5-adb98b094339'}})).respond(201, jasmine.getJSONFixture('GET_api_repository.json'))
      $httpBackend.expect('POST', /\/api\/repositories\/[0-9a-f-]+\/checks/).respond(201, jasmine.getJSONFixture('GET_api_repository_check.json'))

      ctrl.url = 'https://github.com/harrow/frontend.git'
      ctrl.save()
      $httpBackend.flush()

      expect(ctrl.project).toBeDefined()
      expect(ctrl.organization).toBeDefined()
      expect(ctrl.repository).toBeDefined()

      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'quickStart', 'saving')
      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'quickStart', 'completed')
    })

    it('redirects to standard wizard when URL cant determan the names', function () {
      ctrl.url = 'https://github.com/single-repo.git'
      ctrl.save()
      expect(ctrl.$state.go).toHaveBeenCalledWith('wizard.create', {quickStartFailed: true})

      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'quickStart', 'saving')
      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'quickStart', 'error')
    })
  })
})
