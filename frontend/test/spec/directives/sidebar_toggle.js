describe('Directive: sidebarToggle', function () {
  var fixture, el
  beforeEach(angular.mock.inject(function ($rootScope, $compile) {
    fixture = document.createElement('div')
    fixture.id = 'fixture__contents'
    fixture.innerHTML = '<div><div sidebar-toggle>click me</div><div class="sidebar"><a></a></div><div class="empty-space"></div></div>'
    // el = $compile('<div><div sidebar-toggle>click me</div><div class="sidebar"></div><div class="empty-space"></div></div>')($rootScope)
    document.body.appendChild(fixture)
    $compile(fixture)($rootScope)
    fixture = document.querySelector('#fixture__contents')
  }))

  afterEach(function () {
    el = document.body.querySelector('#fixture__contents')
    el.parentElement.removeChild(el)
  })

  it('adds class "sidebar--open" to sidebar when toggle clicked', function () {
    angular.element('[sidebar-toggle]').triggerHandler('click')
    expect(fixture.querySelector('.sidebar').classList).toContain('sidebar--open')
  })

  it('removes class "sidebar--open" when anywhere other than the sidebar is clicked', function () {
    fixture.querySelector('.sidebar').classList.add('sidebar--open')
    angular.element('body').triggerHandler('click')
    expect(fixture.querySelector('.sidebar').classList).not.toContain('sidebar--open')
  })

  it('removes class "sidebar--open" when a anchor is clicked within the sidebar', function () {
    fixture.querySelector('.sidebar').classList.add('sidebar--open')
    angular.element('.sidebar a').triggerHandler('click')
    expect(fixture.querySelector('.sidebar').classList).not.toContain('sidebar--open')
  })
})
