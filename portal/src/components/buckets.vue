<template>
    <div>
        <el-button class="pull-right" type="primary" @click="onAdd">新增</el-button>
        <el-table
                :data="bucketData"
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
                    prop=""
                    label="操作"
                    width="180">
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
        <el-dialog title="编辑bucket" :visible.sync="dialogBucketVisible">
            <div v-if="!isEdit">
                <el-input v-model="newBucketId" placeholder="请输入bucket id"></el-input>
            </div>
            <vue-json-editor v-model="editBucket" :show-btns="false" mode="code"
                             @json-change="onJsonChange"></vue-json-editor>
            <div slot="footer" class="dialog-footer">
                <el-button @click="dialogBucketVisible = false">取 消</el-button>
                <el-button type="primary" @click="onSave(isEdit?editRow.id:newBucketId)">确 定</el-button>
            </div>
        </el-dialog>
    </div>

</template>

<script>
  import vueJsonEditor from 'vue-json-editor'

  export default {
    name: "buckets",
    components: {
      vueJsonEditor
    },
    data() {
      return {
        bucketData: [],
        editBucket: {},
        editRow: {},
        dialogBucketVisible: false,
        isEdit: false,
        newBucketId: null
      }
    },
    methods: {
      loadData() {
        var self = this
        this.$http.get("/api/buckets")
          .then(function (response) {
            self.bucketData = response.data;
          }).catch(function (error) {
          self.$message(error.response.data.message)
        })
      },
      onEdit(row) {
        this.isEdit = true
        this.dialogBucketVisible = true
        this.editRow = row
        this.editBucket = row.basis
      },
      onSave(id) {
        var self = this
        this.dialogBucketVisible = false
        if (this.isEdit) {
          this.$http.put("/api/buckets", {
            "id": id,
            "version": this.editRow.version,
            "basis": this.editBucket
          }).then(function () {
            self.$message("修改成功")
            self.loadData()
          }).catch(function (error) {
            self.$message(error.response.data.message)
          })
        } else {
          this.$http.post("/api/buckets", {
            "id": id,
            "version": "",
            "basis": this.editBucket
          }).then(function () {
            self.$message("创建成功")
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
        this.isEdit = false
        this.dialogBucketVisible = true
        this.editBucket = {}
        this.editRow = {}
      },
      onDelete(row) {
        var self = this
        this.$http.delete(`/api/buckets/${row.id}`)
          .then(function () {
            self.$message("删除成功")
            self.loadData()
          }).catch(function (error) {
          self.$message(error.response.data.message)
        })
      }
    },
    mounted() {
      this.loadData()
    }
  }
</script>
