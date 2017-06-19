describe('Controller: wizardCreateCtrl', function () {
  var $scope, ctrl, $httpBackend, $state, $stateParams, authentication
  beforeEach(angular.mock.inject(function ($rootScope, $controller, _$httpBackend_, _$state_, _authentication_, _$stateParams_) {
    $state = _$state_
    $stateParams = _$stateParams_
    $scope = $rootScope.$new()
    $httpBackend = _$httpBackend_
    authentication = _authentication_
    ctrl = $controller('wizardCreateCtrl', {
      $scope: $scope,
      organization: {
        subject: {
          public: false,
          planUuid: 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3'
        }
      },
      project: {
        subject: {}
      }
    })
    spyOn(ctrl, 'ga')
    spyOn(ctrl.$state, 'go')
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
  }))

  describe('.save()', function () {
    it('sends google analytics', function () {
      $httpBackend.expect('POST', /\/api\/organizations/).respond(201, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('POST', /\/api\/projects/).respond(201, jasmine.getJSONFixture('GET_api_project.json'))

      ctrl.save()
      expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'create', 'formSubmitted')
      $httpBackend.flush()
    })

    it('saves organization then project', function (done) {
      var projectOrg = jasmine.getJSONFixture('GET_api_project_organization.json')
      var project = jasmine.getJSONFixture('GET_api_project.json')

      $httpBackend.expect('POST', /\/api\/organizations/, jasmine.validateHttpParams({subject: {public: false, planUuid: 'b99a21cc-b108-466e-aa4d-bde10ebbe1f3', name: 'ACME Corp.'}})).respond(201, projectOrg)
      $httpBackend.expect('POST', /\/api\/projects/, jasmine.validateHttpParams({subject: { name: 'Meep Meep App'}})).respond(201, jasmine.getJSONFixture('GET_api_project.json'))
      $httpBackend.expect('GET', /\/api\/projects\/[0-9a-f-]+/).respond(200, jasmine.getJSONFixture('GET_api_project.json'))
      ctrl.organization.subject.name = 'ACME Corp.'
      ctrl.project.subject.name = 'Meep Meep App'

      ctrl.save().then(function () {
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'create', 'formSuccess')
        expect(ctrl.project.subject.uuid).toEqual(project.subject.uuid)
        expect($state.go).toHaveBeenCalledWith('wizard.project.connect', {projectUuid: project.subject.uuid}, {reload: true})
        expect(ctrl.flash.success).toEqual('Project Frontend saved')
        done()
      })
      $httpBackend.flush()
    })

    it('saves project if organization is already defined', function (done) {
      $httpBackend.expect('POST', /\/api\/projects/).respond(201, {
        subject: {
          uuid: 'pro123'
        }
      })
      $httpBackend.expect('GET', /\/api\/projects\/pro123/).respond(200, {
        subject: {
          uuid: 'pro123'
        }
      })
      ctrl.project.subject.organizationUuid = 'org123'
      ctrl.save().then(function () {
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'create', 'formSuccess')
        expect(ctrl.project.subject.uuid).toEqual('pro123')
        expect(ctrl.$state.go).toHaveBeenCalledWith('wizard.project.connect', {projectUuid: 'pro123'}, {reload: true})
        done()
      })
      $httpBackend.flush()
    })

    it('catches form error when organization fails', function (done) {
      $httpBackend.expect('POST', /\/api\/organizations/).respond(422)
      ctrl.save().catch(function (reason) {
        expect(ctrl.ga.calls.count()).toEqual(2)
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'create', 'formError')

        done(reason)
      })
      $httpBackend.flush()
    })

    it('catches form error when project fails', function (done) {
      $httpBackend.expect('POST', /\/api\/organizations/).respond(201, {
        subject: {
          uuid: 'org123'
        }
      })
      $httpBackend.expect('POST', /\/api\/projects/).respond(422)
      ctrl.project.subject.name = 'example'
      ctrl.save().catch(function (reason) {
        expect(ctrl.ga.calls.count()).toEqual(2)
        expect(ctrl.ga).toHaveBeenCalledWith('send', 'event', 'wizard', 'create', 'formError')
        expect(ctrl.flash.error).toEqual('Unable to save Project example')

        done(reason)
      })
      $httpBackend.flush()
    })
  })
})
