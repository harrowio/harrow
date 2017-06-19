describe('Controller: breadcrumbsCtrl', function () {
  var ctrl, $controller, $scope, $state, state
  beforeEach(angular.mock.inject(function (_$controller_, $rootScope, _$state_) {
    $controller = _$controller_
    $scope = $rootScope.$new()
    $state = _$state_
    $scope.$resolve = {
      organization: {
        subject: {
          name: 'ACME Corp',
          uuid: 'abc123'
        }
      },
      project: {
        subject: {
          name: 'Meep Meep App',
          uuid: 'abc123'
        }
      },
      script: {
        subject: {
          name: 'Unit Tests',
          uuid: 'abc123'
        }
      },
      task: {
        subject: {
          name: 'Test - Unit Tests'
        }
      },
      operation: {
        subject: {
          createdAt: '2016-04-07T10:59:23.849Z'
        }
      }
    }
  }))
  describe('- script', function () {
    beforeEach(function () {
      state = $state.get('script')
      ctrl = $controller(state.views['header@layout'].controller, {$scope: $scope, $state: {current: state, get: $state.get}})
    })

    it('contains crumbs', function () {
      expect(ctrl.menu[0].name).toEqual('Home')
      expect(ctrl.menu[1]).toEqual(jasmine.objectContaining({title: 'Organization', name: 'ACME Corp'}))
      expect(ctrl.menu[2]).toEqual(jasmine.objectContaining({title: 'Project', name: 'Meep Meep App'}))
      expect(ctrl.menu[3]).toEqual(jasmine.objectContaining({title: 'Script', name: 'Unit Tests'}))
      expect(ctrl.editItem).toEqual(jasmine.objectContaining({stateName: 'script.edit', stateParams: {projectUuid: 'abc123', scriptUuid: 'abc123'}}))
    })
  })

  describe('- createScript', function () {
    beforeEach(function () {
      $scope.$resolve.script.subject.name = ''
      state = $state.get('createScript')
      ctrl = $controller(state.views['header@layout'].controller, {$scope: $scope, $state: {current: state, get: $state.get}})
    })

    it('contains crumbs', function () {
      expect(ctrl.menu[0].name).toEqual('Home')
      expect(ctrl.menu[1]).toEqual(jasmine.objectContaining({title: 'Organization', name: 'ACME Corp'}))
      expect(ctrl.menu[2]).toEqual(jasmine.objectContaining({title: 'Project', name: 'Meep Meep App'}))
      expect(ctrl.menu[3]).toEqual(jasmine.objectContaining({title: '&nbsp;', name: 'New Script'}))
      expect(ctrl.editItem).toBeNull()
    })
  })

  describe('- project', function () {
    beforeEach(function () {
      state = $state.get('projects/show')
      ctrl = $controller(state.views['header@layout'].controller, {$scope: $scope, $state: {current: state, get: $state.get}})
    })

    it('contains crumbs', function () {
      expect(ctrl.menu[0].name).toEqual('Home')
      expect(ctrl.menu[1]).toEqual(jasmine.objectContaining({title: 'Organization', name: 'ACME Corp'}))
      expect(ctrl.menu[2]).toEqual(jasmine.objectContaining({title: 'Project', name: 'Meep Meep App'}))
      expect(ctrl.editItem).toEqual(jasmine.objectContaining({name: 'Project Settings', stateName: 'projects/edit', stateParams: {projectUuid: 'abc123'}}))
    })
  })

  describe('- organization', function () {
    beforeEach(function () {
      delete $scope.$resolve.project
      state = $state.get('organization')
      ctrl = $controller(state.views['header@layout'].controller, {$scope: $scope, $state: {current: state, get: $state.get}})
    })

    it('contains crumbs', function () {
      expect(ctrl.menu[0].name).toEqual('Home')
      expect(ctrl.menu[1]).toEqual(jasmine.objectContaining({title: 'Organization', name: 'ACME Corp'}))
      expect(ctrl.editItem).toEqual(jasmine.objectContaining({name: 'Organization Settings', stateName: 'organization.edit', stateParams: {uuid: 'abc123'}}))
    })
  })

  describe('- operation', function () {
    beforeEach(function () {
      state = $state.get('operations/show')
      ctrl = $controller(state.views['header@layout'].controller, {$scope: $scope, $state: {current: state, get: $state.get}})
    })

    it('contains crumbs', function () {
      expect(ctrl.menu[0].name).toEqual('Home')
      expect(ctrl.menu[1]).toEqual(jasmine.objectContaining({title: 'Organization', name: 'ACME Corp'}))
      expect(ctrl.menu[2]).toEqual(jasmine.objectContaining({title: 'Project', name: 'Meep Meep App'}))
      expect(ctrl.menu[3]).toEqual(jasmine.objectContaining({title: 'Task', name: 'Test - Unit Tests'}))
      expect(ctrl.menu[4]).toEqual(jasmine.objectContaining({title: 'Operation', name: moment('2016-04-07T10:59:23.849Z').format('LLLL')}))
      expect(ctrl.editItem).toBeNull()
    })
  })
})
