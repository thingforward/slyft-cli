var gulp  = require('gulp'),
    zip = require('gulp-zip'),
    runSequence = require('run-sequence'),
    del = require('del'),
    argv = require('yargs').argv,
    exec = require('child_process').exec,
    os = require('os'),
    getos = require('getos');

var pkg = require('./package.json');
var platform = os.platform();
if (platform === "linux") {
  var obj = getos(function(e, os) {
    if (!e) {
      platform = os.dist + '-' + os.release;
      platform = platform.replace(/ /g, '_').toLowerCase();
    }
  });
}

//default task (`gulp`) triggers build
gulp.task('default', ['build']);

gulp.task('build', function(callback) {
  runSequence(
      'clean-bin-dist',
      'go-get',
      'go-fmt',
      'go-vet',
      'go-build',
      'package-binary',
      'dist',
      'clean-bin-home',
      'go-test',
      callback);
});

//call go get without network updates
gulp.task('go-get', function(callback) {
  exec('go get .', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//build but don't install - the end product lives in `dist`
gulp.task('go-build', function(callback) {
  exec('go build .', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//need coverage before adding coverage check
gulp.task('go-test', function(callback) {
  exec('go test .', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//echo required changes, but don't break build
//or modify in-place 
gulp.task('go-fmt', function(callback) {
  exec('gofmt -d *.go', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//why not?
gulp.task('go-vet', function(callback) {
  exec('go vet .', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//at the end, remove the binary that's created by `go build`
gulp.task('clean-bin-home', function() {
  return del.sync(['./slyft-cli', './slyft-cli.exe'], { force: true });
});

//keep only latest version in dist for now - it's not a binrepo
gulp.task('clean-bin-dist', function() {
  return del.sync(['./dist/' + pkg.name + '-*-' + platform + '.zip', './bin/**/*'], { force: true });
});

//move binary to bin, but don't keep it - it's in .gitignore; only *.zips are kept
//whatever the platform, the user calls slyft, not slyft.mac etc.
gulp.task('package-binary', function() {
  return gulp.src(['./slyft-cli', './slyft-cli.exe'], { base: '.' })
  .pipe(gulp.dest('bin'))
});

gulp.task('dist', function() {
  return gulp.src('./bin/**/*', { base: './bin' })
  .pipe(zip(pkg.name + '-' + pkg.version + '-' + platform + '.zip'))
  .pipe(gulp.dest('./dist'));
});

gulp.task('watch', function() {
  gulp.watch(['./*.go'], [
    'build'
  ]);
});
