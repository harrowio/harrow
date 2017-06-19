;(function (window) {
  var cta, scrollTop, ctaTop
  cta = document.querySelector('.blog__callToAction--vertical')
  header = document.querySelector('.layout__header--heavy')
  disable = Math.max(document.body.offsetWidth, document.documentElement.offsetWidth) < 840
  if (!disable && cta) {
    cta.style.width = cta.offsetWidth + 'px'
    window.addEventListener('scroll', function () {
      scrollTop = Math.max(document.body.scrollTop, document.documentElement.scrollTop)
      topLimit = (scrollTop + header.offsetHeight + 10)
      if (topLimit > (ctaTop || cta.offsetTop)) {
        if (!ctaTop) {
          ctaTop = parseInt(cta.offsetTop, 10)
        }
        cta.style.position = 'fixed'
        cta.style.top = header.offsetHeight + 'px'
      } else {
        ctaTop = null
        cta.style.position = null
        cta.style.top = null
      }
    })
  }
})(window)
