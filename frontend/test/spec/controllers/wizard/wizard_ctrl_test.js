describe('Controller: wizardCtrl', function () {
  var $scope, ctrl
  beforeEach(angular.mock.inject(function ($rootScope, $controller) {
    $scope = $rootScope.$new()
    ctrl = $controller('wizardCtrl', {
      $scope: $scope
    })
  }))

  describe('routing', function () {
    var $state, authentication
    beforeEach(angular.mock.inject(function (_$state_, _authentication_) {
      $state = _$state_
      authentication = _authentication_
    }))

    it('redirects to "login" on unauthenticated users', function () {
      $state.go('wizard.create')
      $scope.$digest()
      expect($state.current.name).toEqual('login')
    })

    it('routes wizardCtrl for authenticated users', function () {
      spyOn(authentication, 'hasValidSession').and.callFake(function () {
        return true
      })
      $state.go('wizard.create')
      $scope.$digest()
      expect($state.current.name).toEqual('wizard.create')
      expect($state.current.data.showViews).toEqual([])
      expect($state.current.views['main@wizard'].templateUrl).toEqual('views/wizard/create.html')
      expect($state.current.views['main@wizard'].controller).toEqual('wizardCreateCtrl')
      expect($state.current.views['main@wizard'].controllerAs).toEqual('wizardCreate')
    })
  })

  describe('.menu', function () {
    it('has an array in the correct order', function () {
      angular.mock.inject(function (menuItems) {
        expect(ctrl.menu).toEqual(menuItems.wizard)
      })
    })
  })
})
