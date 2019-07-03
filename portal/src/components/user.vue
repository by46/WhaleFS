<template>
    <el-form ref="form" :model="entity" label-width="140px" label-suffix=":">
        <el-row>
            <el-form-item label="User名称" prop="name">
                <el-input placeholder="User名称"
                          :disabled="isModify"
                          v-model="entity.username"
                          maxlength="128">
                </el-input>
            </el-form-item>
        </el-row>
        <el-row>
            <el-form-item label="Bucket" prop="buckets">
                <el-select v-model="entity.buckets" multiple placeholder="Select"
                           style="width:100%;">
                    <el-option
                            v-for="item in bucketNames"
                            :key="item"
                            :label="item"
                            :value="item">
                    </el-option>
                </el-select>
            </el-form-item>
        </el-row>
        <el-row style="margin-top: 20px;">
            <el-form-item>
                <el-button type="primary" @click="onSave">保存</el-button>
                <el-button @click="onReturn">返回</el-button>
            </el-form-item>
        </el-row>
    </el-form>
</template>

<script>
  export default {
    name: 'user',
    data() {
      return {
        entity: {
          username: '',
          buckets: [],
        },
        isModify: false,
        buckets: [],
        bucketNames: []
      }
    },
    mounted() {
      this.load()
    },
    methods: {
      onReturn() {
        this.$router.push({name: 'users'})
      },
      onSave() {
        let self = this
        if (self.isModify) {
          let body = {
            id: `system.user.${self.entity.username}`,
            basis: this.entity
          }
          self.$http.put(`/api/users`, body)
            .then(() => {
              self.$message.success('修改成功')
              self.load(`system.user.${self.entity.username}`)
            })
            .catch(err => {
              let msg = '修改失败'
              if (err.response) {
                msg = err.response.data.message
              }
              self.$message.error(msg)
            })
        } else {
          let body = {
            id: `system.user.${self.entity.username}`,
            basis: this.entity
          }
          self.$http.post(`/api/users`, body)
            .then(() => {
              self.$message.success('新增成功')
              self.load(`system.user.${self.entity.username}`)
            })
            .catch(err => {
              let msg = '新增失败'
              if (err.response) {
                msg = err.response.data.message
              }
              self.$message.error(msg)
            })
        }
      },
      load(id) {
        let self = this
        let name = id || this.$route.query['id']
        self.$http.get('/api/bucket-names')
          .then(({data}) => {
            self.bucketNames = data
          })
        if (name) {
          self.$http.get(`/api/users/${name}`)
            .then(resp => {
              self.entity.username = resp.data.username
              self.entity.buckets = resp.data.buckets
              self.isModify = true
            }).catch(() => {
            this.$message.error('获取User信息失败')
          })
        }
      }
    }
  }
</script>