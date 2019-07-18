<template>
    <div>
        <el-table
                :data="userData"
                style="width: 100%">
            <el-table-column
                    label="名称"
                    width="200">
                <template slot-scope="{row}">
                    {{row.basis.username}}
                </template>
            </el-table-column>
            <el-table-column
                    label="Buckets">
                <template slot-scope="{row}">
                    <el-tag type="success" v-for="bucket in row.basis.buckets" :key="bucket" style="margin: 3px 3px">
                        {{bucket}}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column
                    label="操作"
                    width="180">
                <template slot="header">
                    <el-button type="primary" @click="onAdd">新增</el-button>
                </template>
                <template slot-scope="{row}">
                    <el-button type="text"
                               @click="onEdit(row)">
                        编辑
                    </el-button>
                    <el-button type="text"
                               style="color: #F56C6C;"
                               @click="onDelete(row)">
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-dialog title="编辑user" :visible.sync="dialogUserVisible">
            <div v-if="!isEdit">
                <el-input v-model="newBucketId" placeholder="请输入user id"></el-input>
            </div>
            <vue-json-editor v-model="editUser" :show-btns="false" mode="code"
                             @json-change="onJsonChange"></vue-json-editor>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogUserVisible = false">取 消</el-button>
                <el-button type="primary" @click="onEdit()">确 定</el-button>
            </div>
        </el-dialog>
    </div>

</template>

<script>
  import vueJsonEditor from 'vue-json-editor'

  export default {
    name: 'users',
    components: {
      vueJsonEditor
    },
    data() {
      return {
        userData: [],
        editUser: {},
        editRow: {},
        dialogUserVisible: false,
        isEdit: false,
        newBucketId: null
      }
    },
    methods: {
      loadData() {
        var self = this
        this.$http.get('/api/users')
        .then(function (response) {
          self.userData = response.data;
        }).catch(function (error) {
          self.$message(error.response.data.message)
        })
      },
      onEdit(row) {
        this.$router.push({name: 'user', query: {id: row.id, version: row.version}})
      },
      onSave(id) {
        var self = this
        this.dialogUserVisible = false
        if (this.isEdit) {
          this.$http.put('/api/users', {
            'id': id,
            'version': this.editRow.version,
            'basis': this.editUser
          }).then(function () {
            self.$message('修改成功')
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        } else {
          this.$http.post('/api/users', {
            'id': id,
            'version': '',
            'basis': this.editUser
          }).then(function () {
            self.$message('创建成功')
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        }
      },
      onJsonChange(value) {
        this.editUser = value
      },
      onAdd() {
        this.$router.push({name: 'user'})
      },
      onDelete(row) {
        let self = this
        this.$confirm('用户删除之后将不能恢复，是否继续', 'Warning', {
          confirmButtonText: '继续',
          cancelButtonText: '取消',
          type: 'warning'
        })
        .then(() => {
          self.$http.delete(`/api/users/${row.id}`)
          .then(function () {
            self.$message('删除成功')
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        })
        .catch(() => {
        })

      }
    },
    mounted() {
      this.loadData()
    }
  }
</script>
