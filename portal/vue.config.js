module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://192.168.1.9:8089',
        changeOrigin: true
      }
    }
  }
}