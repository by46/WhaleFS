<template>
    <el-container>
        <el-header class="header">
            <el-menu
                    style="padding-left: 200px;"
                    :default-active="activeIndex"
                    class="el-menu-demo"
                    @select="handleSelect"
                    mode="horizontal"
                    background-color="#545c64"
                    text-color="#fff"
                    active-text-color="#ffd04b">
                <el-menu-item index="home">Home</el-menu-item>
                <el-menu-item index="buckets">Buckets</el-menu-item>
                <el-menu-item index="users" v-if="username === 'admin'">Users</el-menu-item>
                <el-menu-item index="settings">Settings</el-menu-item>
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
            </el-menu>

        </el-header>
        <el-container>
            <el-main>
                <el-row>
                    <el-col :md="16" :offset="4">
                        <section class="content">
                            <transition
                                    name="page"
                                    mode="out-in">
                                <router-view></router-view>
                            </transition>
                        </section>
                    </el-col>
                </el-row>
            </el-main>
        </el-container>
    </el-container>
</template>

<script>
  export default {
    name: 'portal',
    data() {
      return {
        activeIndex: 'home',
        username: '',
      }
    },
    mounted: function () {
      if (this.$route.name === 'bucket') {
        this.activeIndex = 'buckets'
      } else {
        this.activeIndex = this.$route.name
      }
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
      },
      handleSelect(key) {
        this.$router.push({name: key})
      }
    }
  }
</script>

<style scoped>
    .header {
        line-height: 60px;
        padding: 0;
    }

    .totalUl {
        height: 100%;
    }

    .right-header {
        float: right;
        padding-right: 30px;
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