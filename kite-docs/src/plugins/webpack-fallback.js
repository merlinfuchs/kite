const path = require('path');

module.exports = function (context, options) {
  return {
    name: 'webpack-fallback',
    configureWebpack(config, isServer, utils) {
      return {
        resolve: {
          fallback: {
            buffer: require.resolve('buffer'),
          },
        },
      };
    },
  };
};