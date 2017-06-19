describe "Directive: harrowCan", () ->

  sourceFor = (subject,action) ->
    """
    <harrow-can can-subject="#{subject}" can-action="#{action}">
      <button id="action"></button>
    </harrow-can>
    """

  beforeEach angular.mock.inject (@$compile, @$rootScope, @$timeout) ->
    @$rootScope.project = {
      subject: {},
      _links: {
        self: {
          href: '/projects/:uuid',
        },
        "some-thing": {
          create: 'POST'
          href: '/some-thing'
        }
      }
    }

  it 'parses camelcase', ->
    html = (subject, action) ->
      """
        <div harrow-can="#{action}" can-subject="#{subject}"><span>hello world</span></div>
      """
    el = @$compile(html('project', 'create-someThing'))(@$rootScope)
    @$rootScope.$digest()
    expect(el[0].style.display).toEqual('', 'should be visible')

  describe "when a link rel matching the action is present", () ->
    beforeEach () ->
      @$rootScope.project._links.self.archive = 'DELETE'

    it "shows the element", () ->
      source = sourceFor('project', 'archive')
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      expect(el[0].style.display).not.toEqual("none", "action is visible")

  describe "when no link rel matching the action is present", () ->
    it "hides the element", () ->
      source = sourceFor('project', 'archive')
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      expect(el[0].style.display).toEqual("none", "action is hidden")

  describe "when no links are present", () ->
    beforeEach () ->
     delete @$rootScope.project._links

    it "hides the element", () ->
      source = sourceFor('project', 'archive')
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      expect(el[0].style.display).toEqual("none", "action is hidden")
