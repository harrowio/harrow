const webpackMerge = require('webpack-merge')
const ExtractTextPlugin = require('extract-text-webpack-plugin')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const commonConfig = require('./webpack.common.js')
const helpers = require('./helpers')

module.exports = webpackMerge(commonConfig, {
  devtool: 'cheap-module-eval-source-map',

  output: {
    path: helpers.root('.tmp'),
    publicPath: '/',
    filename: '[name].js',
    chunkFilename: '[id].chunk.js'
  },

  plugins: [
    new ExtractTextPlugin('[name].css'),
    new HtmlWebpackPlugin({
      googleTagManagerId: 'GTM-TD3JKN',
      intercomId: 'h485kip2',
      template: 'app/index.ejs'
    })
  ],

  devServer: {
    quiet: false,
    historyApiFallback: false,
    stats: 'minimal',
    https: false,
    port: 8181,
    inline: true,
    compress: true,
    proxy: {
      '/api/**': {
        target: 'http://localhost:8585/',
        secure: false
      },
      '/ws/**': {
        target: 'http://localhost:8585/',
        secure: false
      }
    }
  }
})


// To point to a live/other instance of the Harrow API server please change the
// `proxy` configuration above to something like this, taking care of the
// pathRewrite option if applicable.
//
// '/api/**': { target: 'https://www.app.harrow.io/', secure: false }
