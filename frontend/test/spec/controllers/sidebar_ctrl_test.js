describe('Controller: SidebarCtrl', function () {
  var $http, $httpBackend, $scope, $ctrl
  var projectUuid = '92cfa771-abcb-4366-3246-846d7a3eac3d'
  var organizationUuid = 'dd2c15ff-cfdf-49ef-3b7d-250be5dc77a7'
  beforeEach(angular.mock.inject(function ($controller, $rootScope, _$http_, _$httpBackend_) {
    $http = _$http_
    $httpBackend = _$httpBackend_
    $scope = $rootScope.$new()
    $ctrl = $controller('sidebarCtrl', {
      $scope: $scope,
      projects: {},
      organizations: {}
    })
    jasmine.authenticate()
  }))

  describe('Organizations', function () {
    it('adds organization after an organization is create', function () {
      $httpBackend.expect('POST', 'http://test.host/api/organizations').respond(201, {
        subject: {
          uuid: organizationUuid,
          name: 'Example'
        }
      })

      $http({
        method: 'POST',
        url: 'http://test.host/api/organizations'
      })

      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('GET_api_user_organizations.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('GET_api_user_projects.json'))
      $httpBackend.flush()

      expect($ctrl.organizationsLength).toEqual(1)
      expect($ctrl.organizationList[organizationUuid].subject.name).toEqual('test')
    })
    it('updates organization after an organization is create', function () {
      $httpBackend.expect('PUT', 'http://test.host/api/organizations/' + organizationUuid).respond(200, {
        subject: {
          uuid: organizationUuid,
          name: 'Example'
        }
      })

      $http({
        method: 'PUT',
        url: 'http://test.host/api/organizations/' + organizationUuid
      })
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('GET_api_user_organizations.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      $httpBackend.flush()
      expect($ctrl.organizationsLength).toEqual(1)
      expect($ctrl.organizationList[organizationUuid].subject.name).toEqual('test')
    })
    it('removes organization after an organization is deleted', function () {
      $httpBackend.expect('DELETE', 'http://test.host/api/organizations/' + organizationUuid).respond(204, '')

      $http({
        method: 'DELETE',
        url: 'http://test.host/api/organizations/' + organizationUuid
      })
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
      $httpBackend.flush()
      expect($ctrl.organizationsLength).toEqual(0)
      expect($ctrl.organizationList[organizationUuid]).toBeUndefined()
    })
  })

  describe('Projects', function () {
    it('adds project after a project has been created', function () {
      $httpBackend.expect('POST', 'http://test.host/api/projects').respond(201, {
        subject: {
          organizationUuid: organizationUuid,
          uuid: projectUuid,
          name: 'hello world'
        },
        _embedded: {
          organizations: [{
            uuid: organizationUuid,
            name: 'Example'
          }]
        }
      })
      $http({
        method: 'POST',
        url: 'http://test.host/api/projects'
      })
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('GET_api_user_organizations.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('GET_api_user_projects.json'))
      $httpBackend.flush()
      expect($ctrl.organizationsLength).toEqual(1)
      expect($ctrl.organizationList[organizationUuid].projects[projectUuid].subject.name).toEqual('test')
    })

    it('updates project after a project has been updated', function () {
      $httpBackend.expect('PUT', 'http://test.host/api/projects/' + projectUuid).respond(200, {
        subject: {
          organizationUuid: organizationUuid,
          uuid: projectUuid,
          name: 'hello world'
        },
        _embedded: {
          organizations: [{
            uuid: organizationUuid,
            name: 'Example'
          }]
        }
      })
      $http({
        method: 'PUT',
        url: 'http://test.host/api/projects/' + projectUuid
      })
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('GET_api_user_organizations.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('GET_api_user_projects.json'))
      $httpBackend.flush()
      expect($ctrl.organizationsLength).toEqual(1)
      expect($ctrl.organizationList[organizationUuid].projects[projectUuid].subject.name).toEqual('test')
    })

    it('removes project after a project has been deleted', function () {
      $ctrl.organizationList = {}
      $ctrl.organizationList[organizationUuid] = {
        uuid: organizationUuid,
        name: 'Example',
        projects: {}
      }
      $ctrl.organizationList[organizationUuid].projects[projectUuid] = {
        uuid: projectUuid,
        name: 'hello world'
      }
      $httpBackend.expect('DELETE', 'http://test.host/api/projects/' + projectUuid).respond(204, '')
      $http({
        method: 'DELETE',
        url: 'http://test.host/api/projects/' + projectUuid
      })
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/organizations/).respond(200, jasmine.getJSONFixture('GET_api_user_organizations.json'))
      $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]*\/projects/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

      $httpBackend.flush()
      expect($ctrl.organizationsLength).toEqual(1)
      expect($ctrl.organizationList[organizationUuid].projects[projectUuid]).toBeUndefined()
    })
  })
})
