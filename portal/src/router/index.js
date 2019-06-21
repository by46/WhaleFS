import Router from 'vue-router'
import Portal from '@/components/layouts/portal'
import Login from '@/components/login'
import Dashboard from '@/components/dashboard'
import Buckets from '@/components/buckets'
import Users from '@/components/users'

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: Portal
    },
    {
      path: '/portal',
      name: 'portal',
      component: Portal,
      children:
        [{
          path: 'dashboard',
          name: 'dashboard',
          component: Dashboard
        }, {
          path: 'buckets',
          name: 'buckets',
          component: Buckets
        }, {
          path: 'users',
          name: 'users',
          component: Users
        }]
    },
    {
      path: '/login',
      name: 'Login',
      component: Login,
    }
  ]
})