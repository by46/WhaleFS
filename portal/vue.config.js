module.exports = {
  devServer: {
    proxy: {
      '/api': {
        target: 'http://10.59.75.71:8089',
        changeOrigin: true
      }
    }
  }
}