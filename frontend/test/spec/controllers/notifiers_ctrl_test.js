describe('Controller: projectNotifiersCtrl', function () {
  var ctrl, $injector
  beforeEach(angular.mock.inject(function ($controller, _$injector_) {
    $injector = _$injector_
    ctrl = $controller('notifiersCtrl', {
      project: {
        subject: {
          uuid: 'abc123'
        }
      },
      notifiers: {},
      scripts: null,
      environments: null,
      tasks: null
    })
  }))

  describe('.menuItems', function () {
    it('has "slack" item', function () {
      expect(ctrl.menuItems.notifier.slackNotifier).toEqual({
        name: 'Slack',
        notifierType: 'slackNotifier',
        icon: 'icon-slack',
        sref: 'notifiers.slackNotifier'
      })
    })

    it('has "email" item', function () {
      expect(ctrl.menuItems.notifier.emailNotifier).toEqual(jasmine.objectContaining({
        name: 'Email',
        notifierType: 'emailNotifier',
        icon: 'icon-mail',
        sref: 'notifiers.emailNotifier'
      }))
    })

    it('has "tasks" item', function () {
      expect(ctrl.menuItems.notifier.taskNotifier).toEqual(jasmine.objectContaining({
        name: 'Task',
        notifierType: 'taskNotifier',
        icon: 'icon-tasks',
        sref: 'notifiers.taskNotifier'
      }))
    })
  })
})
