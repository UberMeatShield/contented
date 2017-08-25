var gulp = require('gulp');
var ts = require('gulp-typescript');
var del = require('del');
var tslint = require('gulp-tslint');
var runSequence = require('run-sequence');
var sourcemaps = require('gulp-sourcemaps');
var webpack = require('webpack-stream');
var ContextReplacementPlugin = require('webpack').ContextReplacementPlugin; //Used to remove bad warnings in angular code
var open = require('gulp-open');
var sass = require('gulp-sass');
var KarmaServer = require('karma').Server;
var path    = require('path');
var shell = require('gulp-shell');
var minify = require('gulp-minify');


var base = './';
var app  = 'contented/';
var dir = {
    base:  base,
    Go: base, // Feels like this should be in a different directory?
    test:  base + 'src/test',
    ts:    base + 'src/ts/',
    sass:  base + 'src/sass/',
    node:  base + 'node_modules/',
    bower: base + 'bower_components/',
    go:    base + app,
    build: base + 'static/build/',
    thirdparty:   base + 'static/third-party'
};

var tasks = {
	defaultTask: 'default',
	typeScript: 'build-ts',
    tslint: 'tslint',
	copy: 'copy',
    copySingleFiles: 'copy-single',
	copyApp: 'copy-app',
	cleanSrc: 'clean-source',
    buildServer: 'build-server',

    buildLib: 'build-lib',
    bundle: 'bundle',
    compress: 'compress',

	watch: 'watch',
	watcherRebuild: 'watcher-rebuild',

    GoWatch: 'Go-watch',
    GoTest: 'Go-test',
    GoChanged: 'Go-changed',
    GoRebuild: 'Go-rebuild',

	buildSass: 'build-sass',
    sassWatch: 'sass-watch',

	tsTest: 'test'
};

// Main task 
gulp.task(tasks.defaultTask, function (cb) {
	runSequence(
        tasks.cleanSrc,
        tasks.buildSass,
		tasks.typeScript,
        tasks.copy,
        tasks.tslint,
        tasks.compress,
        tasks.tsTest,
        tasks.GoWatch,
        tasks.sassWatch,
		tasks.watch
    );
});

gulp.task(tasks.buildServer, function () {
	return runSequence(
        tasks.cleanSrc,
        tasks.buildSass,
		tasks.typeScript,
        tasks.copy,
        tasks.compress
    );
});

// default task starts watcher. in order not to start it each change
// watcher will run the task bellow
gulp.task(tasks.watcherRebuild, function (cb) {
	return runSequence(
		tasks.cleanSrc,
        tasks.tslint,
		tasks.typeScript,
        tasks.copyApp,
        tasks.buildSass,
        tasks.bundle,
        tasks.tsTest,
        cb
    );
});


gulp.task(tasks.tslint, function () {
    return gulp.src(dir.ts + '**/*.ts')
        .pipe(tslint({
            formatter: "verbose",
            configuration: 'config/tslint.json'
        })).pipe(tslint.report());
        
});

// compiles *.ts files by tsconfig.json file and creates sourcemap files
gulp.task(tasks.typeScript, function () {
    console.log("Remove this and just use a webpack task for prod code build");
     var tsProject = ts.createProject('tsconfig.json');
     var tsResult = tsProject.src()
            .pipe(sourcemaps.init())
            .pipe(tsProject());
     return tsResult.js
           .pipe(sourcemaps.write())
           .pipe(gulp.dest(dir.build));
});

gulp.task(tasks.bundle, function() {
    console.log("CHANGE to a webpack config for production");
    var bundleDir = dir.build + 'js/app/boot.js'
    return gulp.src(bundleDir)
        .pipe(webpack({
            module: {
                loaders: [{ loader: 'raw-loader', test: /\.(css|html)$/ }]
			}, 
            plugins: [
            new ContextReplacementPlugin(
                /angular(\\|\/)core(\\|\/)@angular/,
                path.resolve(__dirname, '../src')
            )],
            
            output: {
                filename: 'index.js'
            }
        })).pipe(gulp.dest(dir.build + 'js/'));
});

gulp.task(tasks.compress, function() {
    return gulp.src(dir.build + 'js/index.js')
        .pipe(minify({
            ext: { src: '.js', min: '.min.js'}
        })).pipe(gulp.dest(dir.build + 'js/'));
        
});


gulp.task(tasks.copy, function(cb) {
  runSequence(
    tasks.copyApp,
    tasks.copySingleFiles, 
    tasks.bundle,
    cb
  );
});

// Required to actually run an angluar application
gulp.task(tasks.copySingleFiles, function() {
     return gulp.src([
        dir.node  + 'core-js/client/shim.min.js',
        dir.node + 'zone.js/dist/zone.js'
     ])
    .pipe(gulp.dest(dir.thirdparty));
});


// copy *.html files (templates of components)
// to apropriate directory under public/scripts
gulp.task(tasks.copyApp, function () {
    console.log("I think we can remove this stuff and just use the webpack build for all of it");

	return gulp.src([
      dir.ts + '**/**.html', 
      dir.ts + '**/**.js',
    ]).pipe(gulp.dest(dir.build + '/js'));
});


gulp.task(tasks.buildSass, function () {
	return gulp.src(dir.sass + '/*.scss')
		.pipe(sass())
		.pipe(gulp.dest(dir.build + '/css'));
});

//  clean all generated/compiled files 
//	only in both scripts/ directory
gulp.task(tasks.cleanSrc, function (cb) {
    return del([
         dir.build + '/**/*', 
         dir.ts + 'maps/*'
    ]);
});

// watcher (split into watch sass and watch ts)
gulp.task(tasks.watch, function () {
	gulp.watch([
          dir.ts + '**/**.ts', 
          dir.ts + '**/**.html',
          dir.test + '**/**.ts', 
          './tsconfig.json'
        ], 
        [tasks.watcherRebuild]
    );
});

// Watches the sass, rebuilds (which does the copy) on change
gulp.task(tasks.sassWatch, function() {
    gulp.watch(
      [dir.sass + '**/**.scss'],
      [tasks.buildSass]
    );
});

// Watch our GO files, rebuild the app if they change
gulp.task(tasks.GoWatch, function() {
    gulp.watch(
      [dir.Go + '/**/*.go'],
      [tasks.GoChanged]
    );
});

gulp.task(tasks.GoChanged, function(cb) {
    runSequence(
       //tasks.GoTest, 
       tasks.GoRebuild, // Figure out how to properly run and restart this process
       cb
    );
});

gulp.task(tasks.GoRebuild, shell.task([
      "go build"
    ])
);

gulp.task(tasks.GoTest, shell.task([
      "go test"
    ])
);


gulp.task(tasks.tsTest, function (done) {
	new KarmaServer({
		configFile: __dirname + '/config/karma.conf.js',
        'log-level': 'debug',
		singleRun: true
	}, done).start();
});
