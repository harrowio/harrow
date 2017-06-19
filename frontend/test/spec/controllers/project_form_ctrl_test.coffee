describe "Controller: projectFormCtrl", ->

  beforeEach angular.mock.inject (
    $controller
    @$q
    @$rootScope
  ) =>
    @mockTranslate = jasmine.createSpyObj("$translate", ["instant"])
    @mockState = jasmine.createSpyObj("$state", ["go", "includes"])
    @project =
      subject:
        uuid: "123"
        name: "Test Project"
    @mockProjectResource = {
      save: jasmine.createSpy().and.callFake =>
        @$q.when(@project)
    }
    @controller = $controller "projectFormCtrl",
      $translate: @mockTranslate
      $scope: @scope = $rootScope.$new()
      $state: @mockState
      projectResource: @mockProjectResource
      organization:
        subject:
          uuid: "456"
      project:
        subject:
          organizationUuid: "456"
      members: []

  describe "save", =>
    it "redirects on success", =>
      @controller.project = @project
      @controller.save()
      @$rootScope.$digest()
      expect(@mockState.go).toHaveBeenCalledWith("projects/edit", {uuid: "123"})
