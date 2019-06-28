<template>
    <el-container style="height: 100vh; border: 1px solid #eee">
        <el-header class="header">
            <div class="logo">
                <img src="../../assets/logo.png" alt="whalefs">
            </div>
            <div class="right-header">
                <div v-if="username != ''">
                    <span class="user">当前用户: {{ username }}</span>
                    <el-button style="padding: 0px"
                               type="text"
                               @click="onLogout">
                        退出
                    </el-button>
                </div>
                <div v-else>
                    <el-button style="padding: 0px"
                               type="text"
                               @click="onLogin">
                        登录
                    </el-button>
                </div>
            </div>
        </el-header>

        <el-container>
            <el-aside width="200px" style="background-color: rgb(238, 241, 246)">
                <el-menu :router="true" class="totalUl" :default-active="$route.path">
                    <el-menu-item index="/portal/dashboard"><i class="el-icon-pie-chart"></i>Dashboard</el-menu-item>
                    <el-menu-item index="/portal/buckets"><i class="el-icon-delete"></i>Buckets</el-menu-item>
                    <el-menu-item index="/portal/users"><i class="el-icon-user"></i>Users</el-menu-item>
                </el-menu>
            </el-aside>
            <el-main>
                <section class="content">
                    <transition
                            name="page"
                            mode="out-in">
                        <router-view></router-view>
                    </transition>
                </section>
            </el-main>
        </el-container>
    </el-container>
</template>

<script>
  export default {
    name: 'portal',
    data() {
      return {
        username: '',
      }
    },
    mounted: function () {
      let user = JSON.parse(window.localStorage.getItem('user'))
      this.username = user.username
    },
    methods: {
      onLogout() {
        var self = this
        this.$http.post('/api/logout', {})
        .then(function () {
          window.localStorage.removeItem('user')
          self.$router.push({path: '/login'})
        })
      },
      onLogin() {
        this.$router.push({path: '/login'})
      }
    }
  }
</script>

<style scoped>
    .header {
        color: rgba(255, 255, 255, 0.75);
        line-height: 60px;
        background-color: #24292e;
    }

    .header div {
        display: inline;
    }

    .totalUl {
        height: 100%;
    }

    .right-header {
        float: right;
    }

    .user {
        margin-right: 10px;
    }

    .logo img {
        width: 180px;
        height: 180px;
        margin: -57px 10px -40px -10px;
    }
</style>