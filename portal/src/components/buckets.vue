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
                    {{row.doc.json}}
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
                               type="text">
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
                <el-button type="primary" @click="onSave(isEdit?editId:newBucketId)">确 定</el-button>
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
        editId: null,
        dialogBucketVisible: false,
        isEdit: false,
        newBucketId: null
      }
    },
    methods: {
      loadData() {
        var self = this
        this.axios.get("http://localhost:8089/buckets")
          .then(function (response) {
            self.bucketData = response.data.rows;
          })
      },
      onEdit(row) {
        this.isEdit = true
        this.dialogBucketVisible = true
        this.editBucket = row.doc.json
        this.editId = row.id
      },
      onSave(id) {
        var self = this
        this.dialogBucketVisible = false
        this.axios.post("http://localhost:8089/buckets", {
          "id": id,
          "doc": this.editBucket
        })
          .then(function () {
            self.loadData()
          })
      },
      onJsonChange(value) {
        this.editBucket = value
      },
      onAdd() {
        this.isEdit = false
        this.dialogBucketVisible = true
        this.editBucket = {}
      }
    },
    mounted() {
      this.loadData()
    }
  }
</script>
