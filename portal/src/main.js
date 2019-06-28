import Vue from 'vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'
import axios from 'axios'
import VueAxios from 'vue-axios'

import App from './App.vue'
import router from '@/router'

import 'element-ui/lib/theme-chalk/index.css'

Vue.use(VueRouter)
Vue.use(ElementUI, {size: 'small'})
Vue.use(VueAxios, axios)

Vue.config.productionTip = false
Vue.prototype.BASE_API_URL = "http://localhost:8089"
Vue.prototype.$http.interceptors.request.use(config => {
  let user = JSON.parse(window.localStorage.getItem('user'))

  if (user && user.token) {
    config.headers.Authorization = `Bearer ${user.token}`
  }
  return config
})

Vue.prototype.$http.interceptors.response.use(res => {
  return res
}, error => {
  if (error.response.status === 401) {
    window.localStorage.removeItem('user')
    router.push({path: '/login'})
  }
  return Promise.reject(error);
})

new Vue({
  router,
  render: h => h(App),
}).$mount('#app')
