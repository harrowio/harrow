describe('Controller: dashboardCtrl', function () {
  var ctrl, scope
  beforeEach(angular.mock.inject(function ($controller, $rootScope) {
    scope = $rootScope.$new()
    ctrl = $controller('dashboardCtrl', {
      $scope: scope,
      organizations: [],
      projects: [
        {
          subject: {
            uuid: 'abc123'
          },
          _embedded: {
            organizations: []
          }
        }
      ],
      tasks: [
        {
          subject: {
            projectUuid: 'abc123'
          }
        }
      ],
      cardsByOrganization: []
    })
  }))

  it('has embeded tasks in projects', function () {
    expect(ctrl.projects[0]._embedded.tasks[0]).toBe(ctrl.tasks[0])
  })
})
