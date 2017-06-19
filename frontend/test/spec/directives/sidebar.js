describe('Directive: sidebar', function () {
  var el
  beforeEach(angular.mock.inject(function ($rootScope, $compile) {
    var html = []
    html.push('<div>')
    html.push('  <div class="sidebar">')
    html.push('    <div class="sidebar__header__account">')
    html.push('    </div>')
    html.push('    <div class="sidebar__content">')
    html.push('    </div>')
    html.push('  </div>')
    html.push('</div>')

    el = $compile(html.join(''))($rootScope)
  }))

  it('adds "sidebar__content--second"', function () {
    el.find('.sidebar__header__account').triggerHandler('click')
    expect(el.find('.sidebar__content')[0].classList).toContain('sidebar__content--second')
  })

  it('removes "sidebar__content--second"', function () {
    el.find('.sidebar__content').addClass('sidebar__content--second')
    el.find('.sidebar__header__account').triggerHandler('click')
    expect(el.find('.sidebar__content')[0].classList).not.toContain('sidebar__content--second')
  })
})
