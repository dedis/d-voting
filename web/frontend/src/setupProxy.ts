const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
  if (process.env.NODE_ENV !== 'development') {
    app.use(
      createProxyMiddleware('/api', {
        target: 'http://localhost:5000',
        changeOrigin: true,
        headers: {
          Connection: 'keep-alive',
        },
      })
    );
  }
};
