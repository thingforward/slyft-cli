var gulp  = require('gulp'),
    zip = require('gulp-zip'),
    runSequence = require('run-sequence'),
    del = require('del'),
    argv = require('yargs').argv,
    exec = require('child_process').exec,
    os = require('os'),
    getos = require('getos'),
    md5 = require('gulp-md5');

var pkg = require('./package.json');
var platform = os.platform();
var arch = os.arch();
var execsuffix = "";
if (platform === "linux") {
  var obj = getos(function(e, os) {
    if (!e) {
      platform = os.dist + '-' + os.release;
      platform = platform.replace(/ /g, '_').toLowerCase();
    }
  });
}
if (platform === "win32") {
  execsuffix = ".exe"
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
  exec('go build -o bin/slyft'+execsuffix+' .', function(err, stdout, stderr) {
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
  exec('gofmt -d .', function(err, stdout, stderr) {
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
  return del.sync(['./slyft-cli', './slyft-cli.exe', './slyft', './slyft.exe'], { force: true });
});

//keep only latest version in dist for now - it's not a binrepo
gulp.task('clean-bin-dist', function() {
  return del.sync([
    './dist/' + pkg.name + '-*-' + platform + '_*.zip', //with MD5
    './dist/' + pkg.name + '-*-' + platform + '.zip', //without MD5
    './dist/slyft-*-' + platform + '_*.zip', //with MD5
    './dist/slyft-*-' + platform + '.zip', //without MD5
    './bin/**/*' //original build system
  ], { force: true });
});

//move binary to bin, but don't keep it - it's in .gitignore; only *.zips are kept
//whatever the platform, the user calls slyft, not slyft.mac etc.
gulp.task('package-binary', function() {
  return gulp.src(['./slyft', './slyft.exe'], { base: '.' })
  .pipe(gulp.dest('bin'))
});

gulp.task('dist', function() {
  return gulp.src('./bin/**/*', { base: './bin' })
  .pipe(zip(pkg.name + '-' + pkg.version + '-' + platform + '-' + arch + '.zip'))
  .pipe(md5())
  .pipe(gulp.dest('./dist'));
});

//call this task to cross-compile
gulp.task('build-win32', function(callback) {
  runSequence(
      'go-get',
      'go-fmt',
      'go-vet',
			'go-install-win32',
      'go-build-win32',
      'package-binary',
      'dist',
      'clean-bin-home',
      callback);
});

//install amd64 standard packages
gulp.task('go-install-win32', function(callback) {
  exec('GOOS=windows GOARCH=amd64 go install', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

//now build with hardcoded win32 target
gulp.task('go-build-win32', function(callback) {
	platform = "win32"
	arch = "386"
	execsuffix = ".exe"
  exec('GOOS=windows GOARCH=386 go build -o bin/slyft'+execsuffix+' .', function(err, stdout, stderr) {
    console.log(stdout);
    console.log(stderr);
    callback(err);
  });
});

gulp.task('watch', function() {
  gulp.watch(['./*.go'], [
    'build'
  ]);
});
