import Vue from 'vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'
import axios from 'axios'
import VueAxios from 'vue-axios'

import App from './App.vue'
import router from '@/router'

import 'element-ui/lib/theme-chalk/index.css'

Vue.use(VueRouter)
Vue.use(ElementUI)
Vue.use(VueAxios, axios)

Vue.config.productionTip = false
Vue.prototype.BASE_API_URL = "http://localhost:8089"

new Vue({
    router,
    render: h => h(App),
}).$mount('#app')
