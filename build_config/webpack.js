'use strict';
const path = require('path');
const webpack = require('webpack');

module.exports = {
    entry: {
        'app': './src/ts/app/boot.ts'
    },
    output: {
        filename: '[name].bundle.js'
    },
    // devtool: 'inline-source-map',
    module: {
        rules: [
            {  
                test: /\.json$/,  
                use: [{
                    loader: 'json-loader'
                }]
            },
            {  
                test: /\.(css|html)$/, 
                use: [{
                    loader: 'raw-loader',
                }]
            },
            { 
                test: /\.ts$/, 
                use: [{
                    loader: 'awesome-typescript-loader'
                }]
            }
        ]
    },
    plugins: [
         new webpack.ContextReplacementPlugin(
             /(.+)?angular(\\|\/)core(.+)?/,
             path.resolve(__dirname, '../src')
         )
    ],
    resolve: {
        extensions: ['.js', '.ts', '.json', '*'],
        modules: [
            path.join(__dirname, 'src'), 
            'node_modules'
        ]
    }
};
