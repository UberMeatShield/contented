var gulp = require('gulp');
var ts = require('gulp-typescript');
var del = require('del');
var tslint = require('gulp-tslint');
var runSequence = require('run-sequence');
var sourcemaps = require('gulp-sourcemaps');
var webpack = require('webpack-stream');
var connect = require('gulp-connect');
var open = require('gulp-open');
var sass = require('gulp-sass');
var KarmaServer = require('karma').Server;
var path    = require('path');
var shell = require('gulp-shell');
var uglify = require('gulp-uglify');
var pump = require('pump');
// var cache = require( 'gulp-memory-cache');


var base = './';
var app  = 'contented/';
var dir = {
    base:  base,
    typings: 'typings/',
    test:  base + 'src/test',
    ts:    base + 'src/ts/', 
    sass:  base + 'src/sass/',
    node:  base + 'node_modules/',
    go:    base,
    build: base + 'static/build/',
    thirdparty:   base + 'static/third-party'
};

var tasks = {
    defaultTask: 'default',
    buildDeploy: 'buildDeploy',
    watchers: 'watchers',

    cleanSrc: 'cleanSource',
    watchDoc: 'watchDoc',
    copy: 'copy',
    copySingleFiles: 'copySingleFiles',
    copyLibCSS: 'copyLibCSS',
    copyDocs: 'copyDocs',
    copyFonts: 'copyFonts',

    watchSass: 'watchSass',
    buildSass: 'buildSass',

    watchTypescript: 'watchTypescript',
    tslint: 'tslint',
    typescript: 'typescript', // Webpack bundle operation
    testTypescript: 'testTypescript',
    compress: 'compress',
    rebuildTypescript: 'rebuildTypescript',

    watchGo: 'watchGo',
    rebuildGo: 'rebuildGo',
    changedGo: 'changedGo',
    killGoServer: 'killGoServer',
    serverGo: 'serverGo',
    buildGo: 'buildGo',
    testGo: 'testGo'
};

// Main task 
gulp.task(tasks.defaultTask, [tasks.rebuildTypescript, tasks.rebuildGo, tasks.watchers]);


// Watchers group tasks
gulp.task(tasks.watchers, [tasks.watchGo, tasks.watchDoc, tasks.watchSass, tasks.watchTypescript]);


gulp.task(tasks.buildDeploy, function (callback) {
    runSequence(
        tasks.cleanSrc,
        tasks.copy,
        tasks.buildSass,
        tasks.typescript,
        tasks.compress,
        callback
    );
});

// default task starts watcher. in order not to start it each change
gulp.task(tasks.rebuildTypescript, function(callback) {
    runSequence(
        tasks.cleanSrc, 
        tasks.copy,
        tasks.tslint,
        tasks.buildSass,
        tasks.typescript,
        tasks.compress,
        tasks.testTypescript,
        callback
    );
});

gulp.task(tasks.cleanSrc, function (cb) {
    return del([
         dir.build + '/**/*', 
         dir.ts + 'maps/*'
    ]);
});

// Typescript related tasks
//===================================================
gulp.task(tasks.tslint, function () {
    return gulp.src(dir.ts + '**/*.ts')
        .pipe(tslint({
            formatter: "verbose",
            configuration: 'build_config/tslint.json'
        })).pipe(tslint.report());
});

// watcher (split into watch sass and watch ts)
gulp.task(tasks.watchTypescript, function () {
    gulp.watch([
          dir.ts + '**/**.ts', 
          dir.ts + '**/**.html',
          dir.test + '**/**.ts', 
          './tsconfig.json'
        ], 
        [tasks.rebuildTypescript]
    );
});



gulp.task(tasks.testTypescript, function (done) {
    var singleRun = true; // Running not in single run doesn't seem faster / screws up watchers.
    var karma = new KarmaServer({
        configFile: __dirname + '/build_config/karma.conf.js',
        'log-level': 'error',
        singleRun: singleRun,
    }, done).start();
    return karma;
});


gulp.task(tasks.typescript, function() {
    var bundleEntry = dir.ts + '/app/boot.ts'
    return gulp.src(bundleEntry)
        .pipe(webpack(require('./build_config/webpack.js')))
        .pipe(gulp.dest(dir.build));
});

gulp.task(tasks.compress, function(callback) {
    pump([
        gulp.src(dir.build + 'app.bundle.js'),
        // cache('js'), (did not seem to speed up the process at all)
        uglify(),
        gulp.dest(dir.build + 'min/')
    ], callback);
});


// Initial tasks dealing with copying source. 
// Delete all the gunk out of build directories.
//=================================================
gulp.task(tasks.copy, function(callback) {
  var sequence = runSequence(
    tasks.copyLibCSS,
    tasks.copyFonts,
    tasks.copySingleFiles, 
    tasks.copyDocs,
    callback
  );
  return sequence;
});

gulp.task(tasks.watchDoc, function() {
    gulp.watch(dir.base + 'swagger.yaml', [tasks.copyDocs]);
});

gulp.task(tasks.copyFonts, function() {
    return gulp.src([
        dir.node + 'bootstrap/fonts/*'
    ])
    .pipe(gulp.dest(dir.thirdparty + '/fonts/'));
});

gulp.task(tasks.copyLibCSS, function() {
    return gulp.src([
        dir.node + 'simplemde/dist/simplemde.min.css',
        dir.node + 'bootstrap/dist/css/bootstrap.min.css'
    ])
    .pipe(gulp.dest(dir.thirdparty + '/css/'));
});

gulp.task(tasks.copyDocs, function() {
    //Not async safe, but doesn't really matter since we do no build
    gulp.src([
        dir.base + 'swagger.yaml'
    ]).pipe(gulp.dest(dir.build));
    
    return gulp.src([
      dir.node + 'swagger-ui/dist/**/*'
    ], {base: dir.node}).pipe(gulp.dest(dir.thirdparty));
});

gulp.task(tasks.copySingleFiles, function() {
     return gulp.src([
        dir.node  + 'core-js/client/shim.min.js',
        dir.node + 'zone.js/dist/zone.js'
     ])
    .pipe(gulp.dest(dir.thirdparty));
});


// SASS related operations (does the copy on build)
//=================================================
gulp.task(tasks.watchSass, function() {
    gulp.watch(
      [dir.sass + '**/**.scss'],
      [tasks.buildSass]
    );
});

gulp.task(tasks.buildSass, function () {
    return gulp.src(dir.sass + '/*.scss')
        .pipe(sass())
        .pipe(gulp.dest(dir.build + '/css'));
});


// PYTHON related code sections
//=================================================
gulp.task(tasks.watchGo, function() {
    gulp.watch(
      [dir.go + '/**/*.go',
       dir.base + 'tests/**/*.go'
      ],
      [tasks.changedGo]
    );
});

gulp.task(tasks.changedGo, [tasks.rebuildGo]);

gulp.task(tasks.rebuildGo, function(callback) {
    var restartServer = function(res) { 
        console.log("Waiting for server to die then restarting");
        setTimeout(function() {
            try {
                gulp.src('./contented').pipe(
                    shell(
                      "echo 'Starting up server'; ./contented --dir static/content &"
                    )
                );
            } catch (e) {
                console.error("Failed to glp source anything", e);
            }
            callback(res);
        }, 2000);
    };

    runSequence(
       tasks.killGoServer,
       tasks.buildGo, 
       tasks.testGo,
       restartServer
    );
});

gulp.task(tasks.killGoServer, shell.task([
      "killall contented || echo 'None running' "
    ])
);


gulp.task(tasks.buildGo, shell.task([
      "go build contented"
    ])
);

gulp.task(tasks.testGo, shell.task([
      'echo "Do GO TESTING" '
    ])
);

