import Router from 'vue-router'
import Portal from '@/components/layouts/portal'
import Login from '@/components/login'
import Dashboard from '@/components/dashboard'
import Buckets from '@/components/buckets'
import Bucket from '@/components/bucket'
import Users from '@/components/users'
import User from '@/components/user'
import Settings from '@/components/settings'
import BusUtil from '@/utils/bus'

const router = new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      redirect: '/portal/buckets'
    },
    {
      path: '/portal',
      name: 'portal',
      component: Portal,
      children:
        [
          {
            path: 'home',
            name: 'home',
            component: Dashboard
          },
          {
            path: 'buckets',
            name: 'buckets',
            component: Buckets
          },
          {
            path: 'bucket',
            name: 'bucket',
            component: Bucket
          },
          {
            path: 'users',
            name: 'users',
            component: Users
          },
          {
            path: 'user',
            name: 'user',
            component: User
          },
          {
            path: 'settings',
            name: 'settings',
            component: Settings
          }]
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
    }
  ]
})
router.beforeEach((to, from, next) => {
  BusUtil.load()
  .then(() => {
    next()
  })
})
export default router
