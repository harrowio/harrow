;(function (window) {
  var header, expander
  header = document.querySelector('.layout__header--heavy, .layout__header--hero')
  expander = document.querySelector('.layout__header__expander')
  if (!expander && header) {
    expander = document.createElement('div')
    expander.classList.add('layout__header__expander')
    header.parentNode.insertBefore(expander, header.nextElementSibling)
  } else if (expander) {
    expander.style.height = header.offsetHeight + 'px'
  }

  function stashClass (className) {
    if (header.classList.contains(className) && !header.hasAttribute('data-header-state')) {
      header.setAttribute('data-header-state', className)
      header.classList.remove(className)
    }
  }
  window.addEventListener('scroll', function () {
    var scrollTop = Math.max(document.body.scrollTop, document.documentElement.scrollTop)
    if (scrollTop > 0) {
      stashClass('layout__header--heavy')
      stashClass('layout__header--hero')

      if (!header.classList.contains('layout__header--fixed')) {
        header.classList.add('layout__header--fixed')
        setTimeout(function () {
          header.classList.add('layout__header--isSet')
        }, 50)
      }
    } else {
      if (header.hasAttribute('data-header-state')) {
        header.classList.add(header.getAttribute('data-header-state'))
        header.removeAttribute('data-header-state')
      }
      if (header.classList.contains('layout__header--fixed')) {
        if (scrollTop === 0) {
          header.classList.remove('layout__header--isSet')
          header.classList.remove('layout__header--fixed')
        }
      }
    }
  })
})(window)
