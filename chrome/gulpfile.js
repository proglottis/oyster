'use strict';

var gulp = require('gulp'),
    gulpif = require('gulp-if'),
    argv = require('yargs').argv,
    jshint = require('gulp-jshint'),
    del = require('del'),
    browserify = require('browserify'),
    buffer = require('vinyl-buffer'),
    source = require('vinyl-source-stream'),
    uglify = require('gulp-uglify'),
    eventstream = require('event-stream'),
    ngAnnotate = require('gulp-ng-annotate'),
    sass = require('gulp-sass'),
    concatCss = require('gulp-concat-css');

gulp.task('lint', function() {
  return gulp.src(['./*.js', './app/**/*.js'])
    .pipe(jshint())
    .pipe(jshint.reporter('default'))
    .pipe(jshint.reporter('fail'));
});

gulp.task('clean', function(cb) {
  del(['./dist/*'], cb);
});

gulp.task('javascript', ['lint'], function() {
  var bundles = ['app.js', 'background.js', 'content.js'];

  var tasks = bundles.map(function(filename) {
    return browserify('./app/' + filename).bundle()
      .pipe(source(filename))
      .pipe(buffer())
      .pipe(ngAnnotate())
      .pipe(gulpif(argv.production, uglify()))
      .pipe(gulp.dest('./dist'));
  });
  return eventstream.merge.apply(null, tasks);
});

gulp.task('html', function() {
  return gulp.src('./app/**/*.html')
    .pipe(gulp.dest('./dist'));
});

gulp.task('stylesheet', function() {
  return gulp.src([
      './node_modules/angular/angular-csp.css',
      './app/*.scss'
    ])
    .pipe(sass())
    .pipe(concatCss('app.css'))
    .pipe(gulp.dest('./dist'));
});

gulp.task('image', function() {
  return gulp.src('./app/*.png')
    .pipe(gulp.dest('./dist'));
});

gulp.task('manifest', function() {
  return gulp.src('./app/manifest.json')
    .pipe(gulp.dest('./dist'));
});

gulp.task('build', ['javascript', 'html', 'stylesheet', 'image', 'manifest']);
gulp.task('default', ['build']);
