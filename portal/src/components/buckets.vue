<template>
    <div>
        <el-table
                :data="bucketData"
                stripe="">
            <el-table-column
                    prop="name"
                    label="名称"
                    sortable
                    width="200">
                <template slot-scope="{row}">
                    <el-link :href="bucketHref(row)" type="primary">{{row.basis.name}}</el-link>
                </template>
            </el-table-column>
            <el-table-column
                    prop="basis.memo"
                    sortable
                    label="说明">
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
                    <el-button style="color: #F56C6C;"
                               type="text"
                               @click="onDelete(row)">
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
    </div>

</template>

<script>

  export default {
    name: 'buckets',
    data() {
      return {
        bucketData: [],
        editBucket: {},
        editRow: {},
        isEdit: false,
        newBucketId: null
      }
    },
    computed: {
      bucketHref() {
        return row => {
          let route = this.$router.resolve({name: 'bucket', query: {id: row.id, version: row.version}})
          return route.href
        }
      }
    },
    methods: {
      loadData() {
        var self = this
        this.$http.get('/api/buckets')
        .then(function (response) {
          self.bucketData = response.data;
        }).catch(function (error) {
          self.$message(error.response.data.message)
        })
      },
      onEdit(row) {
        this.$router.push({name: 'bucket', query: {id: row.id, version: row.version}})
      },
      onSave(id) {
        var self = this
        this.dialogBucketVisible = false
        if (this.isEdit) {
          this.$http.put('/api/buckets', {
            'id': id,
            'version': this.editRow.version,
            'basis': this.editBucket
          }).then(function () {
            self.$message('修改成功')
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        } else {
          this.$http.post('/api/buckets', {
            'id': id,
            'version': '',
            'basis': this.editBucket
          }).then(function () {
            self.$message('创建成功')
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        }
      },
      onJsonChange(value) {
        this.editBucket = value
      },
      onAdd() {
        this.$router.push({name: 'bucket'})
      },
      onDelete(row) {
        this.$confirm('Bucket删除之后将不能恢复，是否继续', 'Warning', {
          confirmButtonText: '继续',
          cancelButtonText: '取消',
          type: 'warning'
        })
        .then(() => {
          var self = this
          this.$http.delete(`/api/buckets/${row.id}`)
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
