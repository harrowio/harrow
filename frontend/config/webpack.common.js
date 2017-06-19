const path = require('path')
const webpack = require('webpack')
const autoprefixer = require('autoprefixer')
const ExtractTextPlugin = require('extract-text-webpack-plugin')
const FaviconsWebpackPlugin = require('favicons-webpack-plugin')

const helpers = require('./helpers')

module.exports = {
  entry: {
    app: './app/main.coffee'
  },

  resolve: {
    extensions: ['', '.js', '.ts', '.coffee'],
    alias: {
      jquery: 'jquery/src/jquery'
    }
  },

  module: {
    loaders: [
      {
        test: /\.ts$/,
        loaders: ['ts', 'angular2-template-loader']
      },
      {
        test: /\.coffee$/,
        loader: 'ng-annotate!coffee'
      },
      {
        test: /\.html$/,
        exclude: `${path.resolve(__dirname, '../app/index.html')}`,
        loader: `ngtemplate?relativeTo=${path.resolve(__dirname, '../app')}/!html`
      },
      {
        test: /\.(woff2?|svg|ttf|eot|png|jpg|gif|swf)$/,
        loader: `${require.resolve('file-loader')}?name=[name].[hash].[ext]`
      },
      {
        test: /\.scss$/,
        loader: ExtractTextPlugin.extract('style', 'css?sourceMap&-autoprefixer!postcss-loader!resolve-url!sass?sourceMap')
      },
      {
        test: /\.css$/,
        loader: ExtractTextPlugin.extract('style', 'css?sourceMap')
      }
    ]
  },

  postcss: [
    autoprefixer({
      browsers: ['> 5%', 'last 2 versions']
    })
  ],

  resolveUrlLoader: {
    fail: true,
    absolute: true,
    keepQuery: true
  },

  plugins: [
    new webpack.optimize.CommonsChunkPlugin({
      name: 'app'
    }),
    new FaviconsWebpackPlugin({
      logo: helpers.root('bower_components/style-guide/source/images/harrow logo/harrow-io-icon@2x.png'),
      title: 'Harrow.io',
      background: '#fff',
      icons: {
        android: true,
        appleIcon: true,
        appleStartup: true,
        coast: true,
        favicons: true,
        firefox: true,
        opengraph: true,
        twitter: true,
        windows: true,
        yandex: true
      }
    }),
    new webpack.ProvidePlugin({
      '$': 'jquery',
      'jQuery': 'jquery'
    })
  ]
}
