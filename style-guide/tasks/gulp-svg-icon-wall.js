var fs = require('fs')
var gulp = require('gulp')
var svgmin = require('gulp-svgmin')
var cheerio = require('gulp-cheerio')

function createWall (wall, use) {
  var map = [
    {
      translateX: 22,
      translateY: 6,
      rotate: 330
    },
    {
      translateX: 61,
      translateY: 36,
      rotate: 10
    },
    {
      translateX: 112,
      translateY: 15,
      rotate: 330
    },
    {
      translateX: 182,
      translateY: 16,
      rotate: 20
    },
    {
      translateX: 243,
      translateY: 12,
      rotate: 330
    },
    {
      translateX: 340,
      translateY: 17,
      rotate: 31
    },
    {
      translateX: 409,
      translateY: 21,
      rotate: 314
    },
    {
      translateX: 489,
      translateY: 11,
      rotate: 324
    },
    {
      translateX: 539,
      translateY: 31,
      rotate: 34
    },
    {
      translateX: 6,
      translateY: 57,
      rotate: 20
    },
    {
      translateX: 102,
      translateY: 70,
      rotate: 330
    },
    {
      translateX: 153,
      translateY: 59,
      rotate: 0
    },
    {
      translateX: 215,
      translateY: 60,
      rotate: 330
    },
    {
      translateX: 263,
      translateY: 84,
      rotate: 350
    },
    {
      translateX: 290,
      translateY: 44,
      rotate: 350
    },
    {
      translateX: 366,
      translateY: 59,
      rotate: 324
    },
    {
      translateX: 449,
      translateY: 51,
      rotate: 24
    },
    {
      translateX: 499,
      translateY: 70,
      rotate: 4
    },
    {
      translateX: 559,
      translateY: 80,
      rotate: 324
    },
    {
      translateX: 44,
      translateY: 86,
      rotate: 20
    },
    {
      translateX: 14,
      translateY: 125,
      rotate: 340
    },
    {
      translateX: 74,
      translateY: 126,
      rotate: 0
    },
    {
      translateX: 127,
      translateY: 108,
      rotate: 20
    },
    {
      translateX: 181,
      translateY: 100,
      rotate: 20
    },
    {
      translateX: 175,
      translateY: 145,
      rotate: 330
    },
    {
      translateX: 235,
      translateY: 127,
      rotate: 330
    },
    {
      translateX: 290,
      translateY: 126,
      rotate: 50
    },
    {
      translateX: 320,
      translateY: 86,
      rotate: 310
    },
    {
      translateX: 361,
      translateY: 113,
      rotate: 24
    },
    {
      translateX: 412,
      translateY: 89,
      rotate: 24
    },
    {
      translateX: 405,
      translateY: 141,
      rotate: 64
    },
    {
      translateX: 459,
      translateY: 111,
      rotate: 324
    },
    {
      translateX: 519,
      translateY: 120,
      rotate: 324
    }
  ]

  var translateString = ''
  var el = null
  var offset = 0
  var size = 28
  for (var i = 0; i < map.length; i++) {
    el = use.clone()

    translateString = 'translate(' + (map[i].translateX + offset) + ',' + (map[i].translateY + offset) + ') rotate(' + map[i].rotate + ' ' + (size / 2) + ' ' + (size / 2) + ')'
    el.attr('transform', translateString)
    el.attr('width', size)
    el.attr('height', size)
    wall.append(el)
  }
  return wall
}

gulp.task('svg:icon-wall', function () {
  return gulp.src(['./source/images/icons/*.svg'])
    .pipe(cheerio({
      run: function ($, file) {
        var wrapper = $('svg')
        wrapper.attr('viewBox', '0 0 605 180')
        wrapper.removeAttr('width')
        wrapper.removeAttr('height')
        if (!/^icon-full-color-[\w-]+\.svg$/.test(file.relative)) {
          $('[fill]').removeAttr('fill')
        }
        var name = file.relative.replace('.svg', '')
        var source = $('g#' + name)
        var content = source.clone()
        source.html('')
        createWall(source, content)
      }
    }))
    .pipe(svgmin())
    .pipe(gulp.dest('./source/images/icon-walls/'))
})
