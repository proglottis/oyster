'use strict';

var gulp = require('gulp'),
    jshint = require('gulp-jshint'),
    del = require('del'),
    browserify = require('browserify'),
    transform = require('vinyl-transform'),
    sourcemaps = require('gulp-sourcemaps'),
    uglify = require('gulp-uglify'),
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
  var browserified = transform(function(filename) {
    var b = browserify(filename);
    return b.bundle();
  });

  return gulp.src(['./app/app.js', './app/background.js', './app/content.js'])
    .pipe(browserified)
    .pipe(sourcemaps.init())
      .pipe(uglify())
    .pipe(sourcemaps.write('./'))
    .pipe(gulp.dest('./dist'));
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
