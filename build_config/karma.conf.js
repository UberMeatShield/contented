'use strict';

//browsers: ['Chrome', 'PhantomJS'],
module.exports = function(config) {
    config.set({
        autoWatch: true,
        browsers: ['PhantomJS'],
        files: [
            '../node_modules/core-js/client/shim.js',
            '../node_modules/jquery/dist/jquery.min.js',
            '../build_config/karma.entry.js'
        ],
        frameworks: ['jasmine'],
        logLevel: config.LOG_DEBUG,
        phantomJsLauncher: {
            exitOnResourceError: true
        },
        port: 9876,
        preprocessors: {
            'karma.entry.js': ['webpack', 'sourcemap']
        },
        reporters: ['dots'],
        webpack: require('./webpack.test.js'),
        webpackServer: {
            noInfo: true
        }
    });
};
