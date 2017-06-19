var pkg = require('./package.json'),
  gulp = require('gulp'),
  eol = require('os').EOL,
  fs = require('fs'),
  del = require('del'),
  strip_banner = require('gulp-strip-banner'),
  autoprefixer = require('gulp-autoprefixer'),
  header = require('gulp-header'),
  nodeunit = require('gulp-nodeunit'),
  sass = require('gulp-sass'),
  browserSync = require('browser-sync').create(),
  rename = require('gulp-rename'),
  svgmin = require('gulp-svgmin'),
  svgstore = require('gulp-svgstore'),
  cheerio = require('gulp-cheerio')

require('./tasks/gulp-svg-icon-wall.js')

require('gulp-load')(gulp)
var banner = [ '/** ',
  ' * <%= pkg.name %> - v<%= pkg.version %> - <%= today %>',
  ' * ',
  ' * <%= pkg.author %>, and the web community.',
  ' * Licensed under the <%= pkg.license %> license.',
  ' * ',
  ' * Many thanks to Brad Frost and Dave Olsen for inspiration, encouragement, and advice.',
  ' * ', ' **/'].join(eol)

// load patternlab-node tasks
gulp.loadTasks(__dirname + '/builder/patternlab_gulp.js')

// clean patterns dir
gulp.task('clean', function (cb) {
  del.sync(['./public/patterns/*'], {force: true})
  cb()
})

// build the banner
gulp.task('banner', function () {
  return gulp.src([
    './builder/patternlab.js',
    './builder/object_factory.js',
    './builder/lineage_hunter.js',
    './builder/media_hunter.js',
    './builder/patternlab_grunt.js',
    './builder/patternlab_gulp.js',
    './builder/parameter_hunter.js',
    './builder/pattern_exporter.js',
    './builder/pattern_assembler.js',
    './builder/pseudopattern_hunter.js',
    './builder/list_item_hunter.js',
    './builder/style_modifier_hunter.js'
  ])
    .pipe(strip_banner())
    .pipe(header(banner, {
      pkg: pkg,
    today: new Date().getFullYear() }
    ))
    .pipe(gulp.dest('./builder'))
})

// copy tasks
gulp.task('cp:js', function () {
  return gulp.src('**/*.js', {cwd: './source/js'})
    .pipe(gulp.dest('./public/js'))
})
gulp.task('cp:img', ['svg:store'], function () {
  return gulp.src(
    [ '**/*.svg', '**/*.gif', '**/*.png', '**/*.jpg', '**/*.jpeg'  ],
    {cwd: './source/images'})
    .pipe(gulp.dest('./public/images'))
})
gulp.task('cp:font', function () {
  return gulp.src('**/*', {cwd: './source/fonts'})
    .pipe(gulp.dest('./public/fonts'))
})
gulp.task('cp:data', function () {
  return gulp.src('annotations.js', {cwd: './source/_data'})
    .pipe(gulp.dest('./public/data'))
})

// server and watch tasks
gulp.task('connect', ['lab'], function () {
  browserSync.init({
    server: {
      baseDir: './public/'
    },
    notify: false
  })

  // suggested watches if you use scss
  gulp.watch('./source/js/**/*.js', ['cp:js'])
  gulp.watch('./source/css/**/*.scss', ['sass:style'])
  gulp.watch('./public/styleguide/*.scss', ['sass:styleguide'])
  gulp.watch('./source/images/**/*', ['cp:img'])

  gulp.watch([
    './source/_patterns/**/*.mustache',
    './source/_patterns/**/*.json',
    '!./source/_patterns/00-atoms/03-images/icons.json',
    './source/_data/*.json'	],
    ['lab-pipe'], function () {
      browserSync.reload()
    })

})

// unit test
gulp.task('nodeunit', function () {
  return gulp.src('./test/**/*_tests.js')
    .pipe(nodeunit())
})

// sass tasks, turn on if you want to use
gulp.task('sass:style', function () {
  return gulp.src('./source/css/*.scss')
    .pipe(sass({
      outputStyle: 'expanded',
      precision: 8
    }).on('error', sass.logError))
    .pipe(autoprefixer({
      browsers: [
        'last 2 versions',
        'ie 8',
        'ie 9',
        'android 2.3',
        'android 4',
        'opera 12'
      ]
    }))
    .pipe(gulp.dest('./public/css'))
    .pipe(browserSync.stream())
})
gulp.task('sass:styleguide', function () {
  return gulp.src('./public/styleguide/css/*.scss')
    .pipe(sass({
      outputStyle: 'expanded',
      precision: 8
    }))
    .pipe(gulp.dest('./public/styleguide/css'))
    .pipe(browserSync.stream())
})

// SVG Images
gulp.task('svg:store', function () {
  return gulp.src([
    './source/images/icons/*.svg',
    './source/images/logo-with-type.svg'
  ])
    .pipe(rename(function (path) {
      var name = ['icon']
      if (/^logo-/.test(path.basename)) {
        name = []
      }
      name.push(path.basename)
      path.basename = name.join('-')
    }))
    .pipe(svgmin({
      js2svg: {
        pretty: true
      }
    }))
    .pipe(cheerio({
      run: function ($, file) {
        if (!/^icon-full-color-[\w-]+\.svg$/.test(file.relative)) {
          $('[fill]').removeAttr('fill')
        }
      },
      parserOptions: { xmlMode: true }
    }))
    .pipe(svgstore())
    .pipe(gulp.dest('./source/images/'))
})

gulp.task('svg:doc', ['svg:icon-wall'], function (cb) {
  var iconData = {
    icons: []
  }
  var iconWall = []
  var files = fs.readdirSync('./source/images/icons/')
  var filename = null
  for (var i = 0;i < files.length; i++) {
    if (files[i].indexOf('.svg') >= 0) {
      filename = files[i].replace('.svg', '')
      iconWall.push(filename)
      iconData.icons.push({
        icon: 'icon-' + filename
      })
    }
  }
  fs.writeFileSync('./source/_patterns/00-atoms/03-images/icons.json', JSON.stringify(iconData, null, '  '))
  var varibles = fs.readFileSync('./source/css/generic/_variables.scss', 'utf8')
  varibles = varibles.replace(/^\$harrowIcons: (?:.+)?\;$/gm, '$harrowIcons: ("' + iconWall.join('", "') + '");')
  fs.writeFileSync('./source/css/generic/_variables.scss', varibles, 'utf8')
  cb()
})

gulp.task('lab-pipe', ['lab'], function (cb) {
  cb()
  browserSync.reload()
})

gulp.task('default', ['lab', 'svg:doc'])

gulp.task('assets', ['cp:js', 'cp:img', 'cp:font', 'cp:data', 'sass:style', 'sass:styleguide'])
gulp.task('prelab', ['clean', 'banner', 'assets'])
gulp.task('lab', ['prelab', 'patternlab'], function (cb) { cb()})
gulp.task('patterns', ['patternlab:only_patterns'])
gulp.task('serve', ['lab', 'connect'])
gulp.task('travis', ['lab', 'nodeunit'])

gulp.task('version', ['patternlab:version'])
gulp.task('help', ['patternlab:help'])

/*

*/
