;(function (window) {
  function loadSVG (el, cb) {

    var svgSrc = el.getAttribute('data-svg-src')
    if(!svgSrc) {
      console.warn('Could not find data-svg-src')
      return;
    }
    var xhr = new window.XMLHttpRequest()
    xhr.open('GET', svgSrc)
    xhr.onreadystatechange = function () {
      if (xhr.readyState === 4 && xhr.status === 200) {
        el.innerHTML = xhr.responseText
        if (cb) {
          cb(xhr.responseText)
        }
      }
    }
    xhr.send()
  }

  var svgImages = document.querySelectorAll('[data-svg-src]')
  for (var i = 0; i < svgImages.length; i++) {
    loadSVG(svgImages[i])
  }
})(window)
