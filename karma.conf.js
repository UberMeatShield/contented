// Karma configuration file, see link for more information
// https://karma-runner.github.io/1.0/config/configuration-file.html
var path = require('path');


module.exports = function (config) {

  var tagPath = path.resolve("src/test/mock/tags").replace(new RegExp("/$"), ".json");

  config.set({
    basePath: '',
    frameworks: ['jasmine', '@angular-devkit/build-angular'],
    files: [
      'node_modules/jquery/dist/jquery.min.js',
      { pattern: 'public/static/monaco/**/*.js', watched: false, included: false, served: true },
      { pattern: 'public/static/monaco/**/*.css', watched: false, included: false, served: true },
      { pattern: 'src/test/mock/tags.json', watched: false, included: false, served: true },
    ],
    proxies: {
        '/public/static/monaco/': path.resolve("public/static/monaco/"),
        '/api/tags/': tagPath,
    },
    plugins: [
      require('karma-jasmine'),
      require('karma-chrome-launcher'),
      require('karma-spec-reporter'),
      require('karma-coverage-istanbul-reporter'),
      require('@angular-devkit/build-angular/plugins/karma')
    ],
    client:{
      clearContext: false, // leave Jasmine Spec Runner output visible in browser
      captureConsole: true,
      jasmine: {
          random: false
      }
    },
    coverageIstanbulReporter: {
      reports: [ 'html', 'lcovonly' ],
      fixWebpackSourcePaths: true
    },
    angularCli: {
      environment: 'dev'
    },
    reporters: ['spec'],
    specReporter: {
        maxLogLines: 20,             // limit number of lines logged per test
        suppressErrorSummary: true, // do not print error summary
        suppressFailed: false,      // do not print information about failed tests
        suppressPassed: false,      // do not print information about passed tests
        suppressSkipped: true,      // do not print information about skipped tests
        showSpecTiming: true,      // print the time elapsed for each spec
        failFast: false              // test would finish with error when a first fail occurs.
    },
    port: 9876,
    browserNoActivityTimeout: 4000000,
    colors: true,
    logLevel: config.LOG_DISABLE,
    //logLevel: config.LOG_INFO,
    autoWatch: true,
    browsers: ['ChromeHeadless'],
    singleRun: false
  });
};
