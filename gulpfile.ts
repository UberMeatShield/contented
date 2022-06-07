/**
 *  Gulp is a helper for making the list of commands needed to start the app a bit simpler.
 *  A standard saner dev experience is doing a yarn install then running two gulp commands.
 *
 *  If you do not have global install of gulp, run with "yarn run gulp <cmd>".  Or just global
 *  install it: yarn global add gulp (or npm install -g gulp)
 */
import {src, dest, watch, parallel, series, task} from 'gulp';
import * as cp from 'child_process';
import * as del from 'del';

let base: string = './';
let appName: string = "contented"; // Normal app to build, compile and test
let app: string  = `${appName}/`;
let dir = {
    base:  base,
    test:  base + 'src/test',
    ts:    base + 'src/',
    node:  base + 'node_modules/',
    go:    base,
    deploy: 'public/build/',
    build: base + 'public/build/',
    thirdparty: base + 'public/thirdparty',
};

// Generic execute with a resolve by spawning a child process
const execCmd = (cmd, args, done) => {
    const call = cp.spawn(cmd, args, {stdio: 'inherit'});
    call.on('exit', function(code) {
        if (code !== 0) {
            console.error(`FAILED RUN ${cmd} with ${args}`);
        }
        if (typeof done === 'function') {
            done();
        }
    });
    call.on('disconnect', function() {
        console.log("Lost the parent");
    });
    return call;
};

// Execute a vagrant command inside the vagrant box (keep the vagrant command alone it is a single arg)
const vagrantExec = (vagrantCmd, done) => {
    console.log(`Executing: vagrant ssh -c ${vagrantCmd}`);
    let args = ['ssh', '-t', '-c', vagrantCmd];
    return execCmd('vagrant', args, done);
};

// For running NG commands (we have to split this since it is args on the local box)
const ngExec = (ngCmd, done) => {
    console.log(`Executing: yarn run ${ngCmd}`);
    let args = ngCmd.split(' ');
    args.unshift('run');
    return execCmd('yarn', args, done);
};

// Typescript related build and QA tasks
// ==============================================
const tslint = (done) => {
    return ngExec('ng lint --format=stylish', done);
};

// Build our typescript, the default option is to use Angulars watching cli to only recompile changes
const typescript = (done) => {
    let buildWatch = typeof done === 'function' ? 'false' : 'true';
    let typescriptCmd = `ng build ${appName} --configuration=dev --watch=${buildWatch} --base-href /${dir.deploy}`;
    return ngExec(typescriptCmd, done);
};

// Full tree shaking production build, removes code, minify etc
const typescriptProd = (done) => {
    let cmd = `ng build ${appName} --prod --no-progress --base-href /${dir.deploy}`;
    return ngExec(cmd, done);
};

// This is much faster with monitor=true but harder to properly validate all the steps
// work.   TODO: Determine if it is worth just running in parallel vs debug cost
const typescriptTests = (done) => {
    let testWatch = typeof done === 'function' ? 'false' : 'true';
    let testCmd = `ng test ${appName} --watch=${testWatch}`;
    return ngExec(testCmd, done);
};

// Remove old code that is copied, ensures we are actually building new code
const clean = () => {
    let cleaning = [
         dir.build + '**/*',
         dir.ts + 'maps/*'
    ];
    console.log("Deleting files in: ", cleaning);
    return del(cleaning);
};


// Initial tasks dealing with copying source.
// Delete all the gunk out of build directories.
// =================================================
const copyFonts = async () => {
    return src([]).pipe(dest(dir.thirdparty + '/fonts/'));
};

const copyDocs = async () => {
    return Promise.all([
        src([
            dir.base + 'swagger.yaml'
        ]).pipe(dest(dir.build)),
        src([
            dir.node + 'swagger-ui/dist/**/*'
        ], {base: dir.node}).pipe(dest(dir.thirdparty)),
    ]);
};

const goKillServer = (done) => {
    return execCmd('pkill', ['buffalo'], done);
};

const goBuild = (done) => {
    return execCmd("go", ["build"], done);
};

// Run the tests and use the standard test directory
const goTest = (done) => {
    process.env.DIR = `${__dirname}/mocks/content`;
    execCmd('buffalo', ['test'], done);
};

// Run the buffalo dev server
const goRun = (done) => {
    process.env.DIR = `${__dirname}/mocks/content`;
    execCmd('buffalo', ['dev'], done);
};

// Common group tasks that make up the real watchers and deployment
// ===============================================
const copy = series(copyFonts);  // , copyDocs);
copy.description = "Copy all the various library fonts";

const qa = series(tslint, typescriptTests, goTest);
qa.description = "Run our tests and lint for go and typescript";

const buildDev = series(clean, typescript);
buildDev.description = "Faster no QA version that should get the webapp up and running.";

// Note with most of these changed task there is seperate process running the actual build
const typescriptChanged = series(tslint, typescriptTests, typescript);
typescriptChanged.description = "After a typescript change:  compile, lint and running the tests.";

const buildDeploy = series(clean, typescriptProd);
buildDeploy.description = "The production build, fully compile the typescript / tree shake etc.";

const typescriptWatch = async () => {
    // Kick off a watcher which will run the tests

    typescriptTests(null);

    // Running compile process for the UI Code (self watches)
    typescript(null);

    return watch([
          dir.ts + '**/**.ts',
          dir.ts + '**/**.html',
          dir.test + '**/**.ts',
          './tsconfig.json'
        ],
        series(tslint, copy)
    );
};

// The default task and the most common watching / QA task.
const qaMonitor = series(qa, typescriptWatch);
qaMonitor.description = "Run our QA, then monitor for further changes";

const defaultTasks = series(clean, copy, qaMonitor);
defaultTasks.description = "The standard development watch / build";

// Export all our various tasks
export {
    clean,
    tslint,
    typescript,
    typescriptProd,
    typescriptTests,
    typescriptWatch,
    typescriptChanged,
    copy,
    copyFonts,
    copyDocs,
    goTest,
    goRun,
    goBuild,
    goKillServer,
    qa,
    qaMonitor,
    buildDeploy
};
export default defaultTasks;



