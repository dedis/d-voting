const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function (app) {
    app.use(
        createProxyMiddleware('/evoting', {
            target: 'http://localhost:5000', // API endpoint 1
            changeOrigin: true,
            headers: {
                Connection: "keep-alive"
            }
        })
    );
    app.use(
        createProxyMiddleware('/api', {
            target: 'http://localhost:5000', // API endpoint 2
            changeOrigin: true,
            headers: {
                Connection: "keep-alive"
            }
        })
    );
}