const path = require('path')
const webpack = require('webpack')
const HtmlWebpackPlugin = require('html-webpack-plugin')
const helpers = require('./helpers')

module.exports = {
  devtool: 'inline-source-map',

  resolve: {
    extensions: ['', '.ts', '.js', '.coffee']
  },

  module: {
    loaders: [
      {
        test: /\.js$/,
        loader: 'babel',
        exclude: /(node_modules|bower_components|braintree\.js|lom_bundle\.js)/
      },
      {
        test: /\.ts$/,
        loaders: ['ts', 'angular2-template-loader']
      },
      {
        test: /icons\.svg$/,
        loader: 'file?name=images/icons.svg'
      },
      {
        test: /\.coffee$/,
        loader: 'coffee'
      },
      {
        test: /\.html$/,
        loader: `ngtemplate?relativeTo=${path.resolve(__dirname, '../app')}/!html`
      },
      {
        test: /\.json$/,
        loader: 'json'
      },
      {
        test: /\.(png|jpe?g|gif|svg|woff|woff2|ttf|eot|ico|swf|s?css)$/,
        loader: 'null',
        exclude: [/icons\.svg/]
      },
      {
        test: 'jquery\.js',
        loader: 'expose?jQuery'
      }
    ]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: 'app/index.ejs'
    }),
    new webpack.ProvidePlugin({
      '$': 'jquery',
      'jQuery': 'jquery'
    })
  ]
}
