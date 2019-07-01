<template>
    <div>
        <el-table
                :data="userData"
                style="width: 100%">
            <el-table-column
                    prop="id"
                    label="ID"
                    width="180">
            </el-table-column>
            <el-table-column
                    prop="doc"
                    label="内容">
                <template slot-scope="{row}">
                    {{row.basis}}
                </template>
            </el-table-column>
            <el-table-column
                    label="操作"
                    width="180">
                <template slot="header">
                    <el-button type="primary" @click="onAdd">新增</el-button>
                </template>
                <template slot-scope="{row}">
                    <el-button style="padding: 0px"
                               type="text"
                               @click="onEdit(row)">
                        编辑
                    </el-button>
                    <el-button style="padding: 0px"
                               type="text"
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
                <el-button type="primary" @click="onSave(isEdit?editRow.id:newBucketId)">确 定</el-button>
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
        this.isEdit = true
        this.dialogUserVisible = true
        this.editRow = row
        this.editUser = row.basis
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
        this.isEdit = false
        this.dialogUserVisible = true
        this.editUser = {}
        this.editRow = {}
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
