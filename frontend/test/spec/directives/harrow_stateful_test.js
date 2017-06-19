describe('Directive: harrowStateful', function () {
  var $compile, stateful, $scope
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope, Stateful) {
    stateful = new Stateful()
    $compile = _$compile_
    $scope = $rootScope
  }))

  describe('initial state', function () {
    it('transitions states with uiSref and harrowCan', function () {
      $scope.project = {
        _links: {
          self: {
            href: 'http://test.host',
            create: 'POST',
            archive: 'DELETE'
          }
        }
      }
      $scope.options = {}
      stateful.on('pending', function () {
        $scope.options = {
          content: 'pending state'
        }
      })
      stateful.on('complete', function () {
        $scope.options = {
          content: 'Completed',
          attrs: {
            canSubject: 'project',
            canAction: 'archive',
            uiSref: 'dashboard'
          }
        }
      })
      var html = '<a harrow-stateful="options" harrow-can can-action="create" can-subject="project">initial state</a>'
      var el = $compile(html)($scope)
      $scope.$digest()

      expect(el[0].textContent).toEqual('initial state', 'intial state defined in html')
      expect(el[0].getAttribute('can-action')).toEqual('create')

      stateful.transitionTo('pending')
      $scope.$digest()
      expect(el[0].textContent).toEqual('pending state', 'changes text content to pending')
      stateful.transitionTo('complete')

      $scope.$digest()
      expect(el[0].textContent).toEqual('Completed', 'changes text content to final state')
      expect(el[0].getAttribute('can-action')).toEqual('archive')
      expect(el[0].getAttribute('href')).toEqual('#/a/dashboard', 'ui-sref should compiles href')
      expect(el[0].style.display).not.toEqual('none', 'hidden')
    })

    it('parsed options', function () {
      $scope.obj = {}
      var html = '<a harrow-stateful="obj">initial state</a>'
      var el = $compile(html)($scope)
      expect(el.text()).toEqual('initial state')

      var obj = {
        state: 'completed',
        content: 'Completed'
      }
      $scope.obj = obj

      $scope.$digest()
      expect(el.text()).toEqual('Completed')
    })

    it('should not persist initial ngClick event', function () {
      $scope.onInitialClick = function () {
        $scope.initialClicked = true
      }
      $scope.onCompletedClick = function () {
        $scope.completedClicked = true
      }
      $scope.statefulOptions = {
        content: 'Clicked',
        attrs: {
          ngClick: 'onCompletedClick()'
        }
      }
      var html = '<a harrow-stateful="statefulOptions" ng-click="onInitialClick()">initial state</a>'
      var el = $compile(html)($scope)
      $scope.$digest()
      $scope.completed = true
      $scope.$digest()
      el.triggerHandler('click')
      expect($scope.initialClicked).not.toBeTruthy()
      expect($scope.completedClicked).toBeTruthy()
    })
  })
})
