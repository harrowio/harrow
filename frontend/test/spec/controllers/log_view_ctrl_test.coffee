xdescribe "Controller: logViewCtrl", ->

  beforeEach angular.mock.inject ($controller, $rootScope, @$q) =>
    window.Lom =
      Lom: () -> jasmine.createSpyObj('lom', ['pushEvent', 'render'])
      defaultHandlers: () -> []

    @scope = $rootScope.$new()

    @controller = $controller "logViewCtrl", {
      $scope: @scope
      $element:
        find: jasmine.createSpy('$element.find').and.returnValue([0])
      $state: @mockState
      ws: jasmine.createSpyObj('ws', ['subLogevents'])
      $window: jasmine.createSpyObj('window', ['document'])
      $timeout: jasmine.createSpy('$timeout')
    }
    @scope.ctrl = @controller

  it "follows logs by default", =>
    expect(@controller.isFollowing()).toBe true

  describe "when toggling the logs", =>
    it  "does not follow the logs anymore", =>
      @controller.toggleFollowing()
      expect(@controller.isFollowing()).toBe false

  describe "when following the logs", =>
    it "calls follow", () =>
      @controller.follow = jasmine.createSpy('follow')
      @controller.handleLogEvent({})
      expect(@controller.follow).toHaveBeenCalled()

  describe "when not following", =>
    beforeEach () =>
      @controller.toggleFollowing() if @controller.isFollowing()

    it "does not call follow", =>
      @controller.follow = jasmine.createSpy('follow')
      @controller.handleLogEvent({})
      expect(@controller.follow).not.toHaveBeenCalled()
