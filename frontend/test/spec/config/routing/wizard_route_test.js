describe('Routing: wizard', function () {
  var $scope, $state, $httpBackend, authentication
  beforeEach(angular.mock.inject(function ($rootScope, _$state_, _$httpBackend_, _authentication_, userResource) {
    $httpBackend = _$httpBackend_
    $scope = $rootScope
    $state = _$state_
    authentication = _authentication_
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
    spyOn(authentication, 'hasNoSession').and.callFake(function () {
      return false
    })

    jasmine.authenticate()
  }))

  it("redirects to wizard, when org's are empty", function () {
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/organizations/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/projects/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/organizations/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

    $state.go('dashboard')
    $httpBackend.flush()

    expect($state.current.name).toEqual('wizard.quick-start')
  })

  it('passes through to wizard', function () {
    var state = $state.get('wizard')

    expect(state.parent).toEqual('layout')
    expect(state.name).toEqual('wizard')
    expect(state.url).toEqual('/a/wizard')
    expect(state.data.requiresAuth).toBeTruthy()

    angular.mock.inject(function ($controller, menuItems) {
      var ctrl = $controller(state.views['app-sidebar@layout'].controller, {$scope: $scope, task: null})
      expect(ctrl.menu).toEqual(menuItems.wizard)
    })

    expect(state.views['app-sidebar@layout'].controllerAs).toEqual('appSidebar')
  })

  it('redirects to "login" on unauthenticated users', function () {
    authentication.hasValidSession.and.callFake(function () {
      return false
    })
    authentication.hasNoSession.and.callFake(function () {
      return true
    })
    $state.go('wizard.create')
    $scope.$digest()
    expect($state.current.name).toEqual('login')
  })

  it('routes wizardCtrl for authenticated users', function () {
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/organizations/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/projects/).respond(200, jasmine.getJSONFixture('empty_collection.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+\/jobs/).respond(200, jasmine.getJSONFixture('empty_collection.json'))

    $state.go('wizard.create')
    $httpBackend.flush()
    expect($state.current.name).toEqual('wizard.create')
    expect($state.current.data.showViews).toEqual([])
    expect($state.current.views['main@wizard'].templateUrl).toEqual('views/wizard/create.html')
    expect($state.current.views['main@wizard'].controller).toEqual('wizardCreateCtrl')
    expect($state.current.views['main@wizard'].controllerAs).toEqual('wizardCreate')
  })
})
