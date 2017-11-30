var gulp = require('gulp');
var del = require('del');
var runSequence = require('run-sequence');
var sass = require('gulp-sass');
var KarmaServer = require('karma').Server;
var shell = require('gulp-shell');

var base = './';
var app  = 'contented/';
var dir = {
    base:  base,
    typings: 'typings/',
    test:  base + 'src/test',
    ts:    base + 'src/', 
    sass:  base + 'src/scss/',
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
    typescript: 'typescript', 
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
gulp.task(tasks.watchers, [tasks.watchTypescript, tasks.watchGo, tasks.watchDoc, tasks.watchSass]);


gulp.task(tasks.buildDeploy, function (callback) {
    runSequence(
        tasks.cleanSrc,
        tasks.typescript,
        tasks.copy,
        tasks.buildSass,
        callback
    );
});

// default task starts watcher. in order not to start it each change
gulp.task(tasks.rebuildTypescript, function(callback) {
    runSequence(
        tasks.cleanSrc, 
        tasks.tslint,
        tasks.typescript,
        tasks.buildSass,
        tasks.copy,
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
// watcher (split into watch sass and watch ts)
gulp.task(tasks.watchTypescript, function () {
    gulp.watch([
          dir.ts + '**/**.ts', 
          dir.ts + '**/**.html',
          dir.test + '**/**.ts'
        ], 
        [tasks.rebuildTypescript]
    );
});

gulp.task(tasks.tslint, shell.task([
    'ng lint'
]));

gulp.task(tasks.testTypescript, shell.task([
    'ng test --env dev'
]));

gulp.task(tasks.typescript, shell.task([
    'ng build --env dev --deploy-url /' + dir.build
]));

gulp.task(tasks.compress, shell.task([
    'ng build --prod --deploy-url /' + dir.build
]));

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
                      "echo 'Starting up server'; ./contented --dir static/content/ &"
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
      "killall -9 contented || echo 'None running' "
    ])
);

gulp.task(tasks.buildGo, shell.task([
      "go build contented"
    ])
);

gulp.task(tasks.testGo, shell.task([
      'go test $(go list ./... | grep -v /vendor/)'
    ])
);

