const webpack = require('webpack')
const webpackMerge = require('webpack-merge')
const ExtractTextPlugin = require('extract-text-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const commonConfig = require('./webpack.common.js')
const helpers = require('./helpers')

const ENV = process.env.NODE_ENV = process.env.ENV = process.env.HARROW_ENV || 'production'

console.log('Building Harrow.io Production Mode: ENV=%s', ENV)

module.exports = webpackMerge(commonConfig, {
  devtool: 'source-map',

  output: {
    path: helpers.root('dist'),
    publicPath: '/',
    filename: '[name].[hash].js',
    chunkFilename: '[id].[hash].chunk.js'
  },

  htmlLoader: {
    minimize: false // workaround for ng2
  },

  plugins: [
    new webpack.NoErrorsPlugin(),
    new webpack.optimize.DedupePlugin(),
    new webpack.optimize.UglifyJsPlugin({
      mangle: false
    }),
    new ExtractTextPlugin('[name].[hash].css'),
    new webpack.DefinePlugin({
      'process.env': {
        'ENV': JSON.stringify(ENV)
      }
    }),
    new HtmlWebpackPlugin({
      googleTagManagerId: ENV == 'enterprise' ? '' : 'GTM-TD3JKN',
      intercomId: ENV == 'enterprise' ? '' : 'og7d4w8q',
      template: 'app/index.ejs'
    })
  ]
})
