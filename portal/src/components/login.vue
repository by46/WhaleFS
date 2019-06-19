<template>
    <div class="login-container">
        <el-form ref="loginForm" :model="loginForm" :rules="loginRules" class="login-form" autocomplete="on" label-position="left">

            <div class="title-container">
                <h3 class="title">Login Form</h3>
            </div>

            <el-form-item prop="username">
        <span class="svg-container">
            <i class="el-icon-s-custom"></i>
        </span>
                <el-input
                        ref="username"
                        v-model="loginForm.username"
                        placeholder="Username"
                        name="username"
                        type="text"
                        tabindex="1"
                        autocomplete="on"
                />
            </el-form-item>

            <el-tooltip v-model="capsTooltip" content="Caps lock is On" placement="right" manual>
                <el-form-item prop="password">
          <span class="svg-container">
            <i class="el-icon-s-goods"></i>
          </span>
                    <el-input
                            :key="passwordType"
                            ref="password"
                            v-model="loginForm.password"
                            :type="passwordType"
                            placeholder="Password"
                            name="password"
                            tabindex="2"
                            autocomplete="on"
                            @keyup.native="checkCapslock"
                            @blur="capsTooltip = false"
                            @keyup.enter.native="handleLogin"
                    />
                </el-form-item>
            </el-tooltip>

            <el-button :loading="loading" type="primary" style="width:100%;margin-bottom:30px;" @click.native.prevent="handleLogin">Login</el-button>
        </el-form>
    </div>
</template>

<script>
  export default {
    name: 'Login',
    data() {
      const validatePassword = (rule, value, callback) => {
        if (value.length < 6) {
          callback(new Error('The password can not be less than 6 digits'))
        } else {
          callback()
        }
      }
      return {
        loginForm: {
          username: '',
          password: ''
        },
        loginRules: {
          username: [{ required: true, trigger: 'blur'}],
          password: [{ required: true, trigger: 'blur', validator: validatePassword }]
        },
        passwordType: 'password',
        capsTooltip: false,
        loading: false,
        showDialog: false,
        redirect: undefined,
        otherQuery: {}
      }
    },
    mounted() {
      if (this.loginForm.username === '') {
        this.$refs.username.focus()
      } else if (this.loginForm.password === '') {
        this.$refs.password.focus()
      }
    },
    methods: {
      checkCapslock({ shiftKey, key } = {}) {
        if (key && key.length === 1) {
          if (shiftKey && (key >= 'a' && key <= 'z') || !shiftKey && (key >= 'A' && key <= 'Z')) {
            this.capsTooltip = true
          } else {
            this.capsTooltip = false
          }
        }
        if (key === 'CapsLock' && this.capsTooltip === true) {
          this.capsTooltip = false
        }
      },
      handleLogin() {
        this.$refs.loginForm.validate(valid => {
          if (valid) {
            this.loading = true
            this.axios.post(this.BASE_API_URL + '/api/login', this.loginForm)
              .then((resp) => {
                window.localStorage.setItem('user', JSON.stringify(resp.data))
                this.$router.push({ path: '/portal', query: this.otherQuery })
                this.loading = false
              })
              .catch(() => {
                this.loading = false
              })
          } else {
            return false
          }
        })
      }
    }
  }
</script>

<style>
    /* 修复input 背景不协调 和光标变色 */
    /* Detail see https://github.com/PanJiaChen/vue-element-admin/pull/927 */
    @supports (-webkit-mask: none) and (not (cater-color: $cursor)) {
        .login-container .el-input input {
            color: #fff;
        }
    }
    /* reset element-ui css */
    .login-container .el-input {
        display: inline-block;
        height: 47px;
        width: 85%;
    }
    .login-container .el-input input {
        background: transparent;
        border: 0px;
        -webkit-appearance: none;
        border-radius: 0px;
        padding: 12px 5px 12px 15px;
        color: #fff;
        height: 47px;
        caret-color: #fff;
    }
    .login-container .el-input input:-webkit-autofill {
        box-shadow: 0 0 0px 1000px #283443 inset !important;
        -webkit-text-fill-color: #fff !important;
    }
    .login-container .el-form-item {
        border: 1px solid rgba(255, 255, 255, 0.1);
        background: rgba(0, 0, 0, 0.1);
        border-radius: 5px;
        color: #454545;
    }
</style>

<style scoped>
    .login-container {
        min-height: 100%;
        width: 100%;
        background-color: #2d3a4b;
        overflow: hidden;
    }
    .login-container .login-form {
        position: relative;
        width: 520px;
        max-width: 100%;
        padding: 160px 35px 0;
        margin: 0 auto;
        overflow: hidden;
    }
    .login-container .tips {
        font-size: 14px;
        color: #fff;
        margin-bottom: 10px;
    }
    .login-container .tips span:first-of-type {
        margin-right: 16px;
    }
    .login-container .svg-container {
        padding: 6px 5px 6px 15px;
        color: #889aa4;
        vertical-align: middle;
        width: 30px;
        display: inline-block;
    }
    .login-container .title-container {
        position: relative;
    }
    .login-container .title-container .title {
        font-size: 26px;
        color: #eee;
        margin: 0px auto 40px auto;
        text-align: center;
        font-weight: bold;
    }
    .login-container .show-pwd {
        position: absolute;
        right: 10px;
        top: 7px;
        font-size: 16px;
        color: #889aa4;
        cursor: pointer;
        user-select: none;
    }
    .login-container .thirdparty-button {
        position: absolute;
        right: 0;
        bottom: 6px;
    }
    @media only screen and (max-width: 470px) {
        .login-container .thirdparty-button {
            display: none;
        }
    }

    .svg-icon {
        width: 1em;
        height: 1em;
        vertical-align: -0.15em;
        fill: currentColor;
        overflow: hidden;
    }

</style>