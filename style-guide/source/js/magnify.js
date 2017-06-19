;(function () {
  function getNativeDimentionsFor (img) {
    var originalImg = new Image()
    originalImg.src = img.src
    originalImg.onload = function () {}
    return {
      width: originalImg.width,
      height: originalImg.height
    }
  }

  function setMagnifyGlass (img) {
    var dims = getNativeDimentionsFor(img)
    var magnifier = img.parentElement
    var glass = magnifier.querySelector('span')
    if (!glass) {
      glass = document.createElement('span')
    }
    glass.style.backgroundImage = 'url(' + img.src + ')'
    glass.classList.add('magnify__glass')

    var data = {}
    data.x = parseInt(magnifier.getAttribute('data-x'), 10) || 0
    data.y = parseInt(magnifier.getAttribute('data-y'), 10) || 0
    data.ratio = magnifier.getAttribute('data-ratio') || 1

    data.xx = data.x / (dims.width / img.offsetWidth)
    data.yy = data.y / (dims.height / img.offsetHeight)

    if (data.x > 0 && data.y > 0) {
      magnifier.appendChild(glass)
    }

    var bgOffset = {}
    bgOffset.x = Math.round((data.xx - glass.offsetLeft) / img.offsetWidth * (dims.width / data.ratio) - glass.offsetWidth / 2) * -1
    bgOffset.y = Math.round((data.yy - glass.offsetTop) / img.offsetHeight * (dims.height / data.ratio) - glass.offsetHeight / 2) * -1
    glass.style.backgroundPosition = bgOffset.x + 'px ' + bgOffset.y + 'px'
    glass.style.backgroundSize = dims.width / data.ratio + 'px ' + dims.height / data.ratio + 'px'
    glass.style.left = data.xx - glass.offsetWidth / 2 + 'px'
    glass.style.top = data.yy - glass.offsetHeight / 2 + 'px'
  }

  function enableMagnifier () {
    var imgs = document.querySelectorAll('.magnify img')
    for (var i = 0; i < imgs.length; i++) {
      setMagnifyGlass(imgs[i])
    }
  }
  enableMagnifier()
  document.addEventListener('DOMContentLoaded', enableMagnifier, false)
})(window)
