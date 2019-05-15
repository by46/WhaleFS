import Router from 'vue-router'
import Portal from '@/components/layouts/portal'
import Dashboard from '@/components/dashboard'

export default new Router({
    mode: 'history',
    routes: [{
        path: '/portal',
        name: 'portal',
        component: Portal,
        children: [{
            path: '',
            name: 'dashboard',
            component: Dashboard
        }]
    }]
})