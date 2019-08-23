<template>
    <div>
        <el-divider content-position="left">App ID/App Secret Key</el-divider>
        <el-table :data="accessKeys" stripe>
            <el-table-column type="expand">
                <template slot-scope="props">
                    <el-row>
                        <el-col :md="1">Scope:</el-col>
                        <el-col :md="23">
                            <el-checkbox-group v-model="scope">

                                <el-checkbox :label="bucket.basis.name" v-for="bucket in buckets"></el-checkbox>
                            </el-checkbox-group>
                        </el-col>
                    </el-row>
                </template>
            </el-table-column>
            <el-table-column label="创建时间"
                             prop="create_date"
                             width="100px"
                             sortable>
                <template slot-scope="{row}">
                    {{row.create_date | lte-date}}
                </template>
            </el-table-column>
            <el-table-column label="AccessKey/SecretKey" width="410px">
                <template slot-scope="{row}">
                    <el-row>
                        <el-col :md="2">AK:</el-col>
                        <el-col :md="22">
                            <el-input v-model="row.app_key" :disabled="true"></el-input>
                        </el-col>
                    </el-row>
                    <el-row style="margin-top: 5px;">
                        <el-col :md="2">SK:</el-col>
                        <el-col :md="22">
                            <lte-input v-model="row.secret_key" :disabled="true" show-password></lte-input>
                        </el-col>
                    </el-row>
                </template>
            </el-table-column>
            <el-table-column label="过期时间" sortable prop="expires">
                <template slot-scope="{row}">
                    <el-date-picker
                            style="width: 100%;"
                            v-model="row.expires"
                            type="date"
                            placeholder="过期时间"
                            value-format="timestamp"
                            :picker-options="pickerOptions">
                    </el-date-picker>
                </template>
            </el-table-column>
            <el-table-column label="状态"
                             prop="enable"
                             width="100px"
                             sortable>
                <template slot-scope="{row}">
                    {{row.enable | lte-access-status}}
                </template>
            </el-table-column>
            <el-table-column label="操作" width="150px">
                <template slot-scope="{row}">
                    <el-button @click="onChangeStatus(row, false)"
                               v-if="row.enable"
                               type="text">禁用
                    </el-button>
                    <el-button @click="onChangeStatus(row, true)" v-else
                               type="text">启用
                    </el-button>
                    <el-button @click="onUpdate(row)" type="text">保存</el-button>
                    <el-button style="color: #F56C6C;"
                               type="text"
                               @click="onDelete(row)">
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-row style="margin-top: 10px" type="flex" justify="end">
            <el-button type="primary" @click="onCreate">创建Access Key/Secret Key</el-button>
        </el-row>
    </div>
</template>

<script>
  import _ from 'lodash'

  import LteInput from './lte-input'

  export default {
    name: 'settings',
    components: {LteInput},
    filters: {
      'lte-access-status': (value) => {
        return value ? '使用中' : '已停用'
      }
    },
    data() {
      return {
        pickerOptions: {
          disabledDate(time) {
            return time.getTime() < Date.now();
          }
        },
        accessKeys: [],
        buckets: []
      }
    },
    methods: {
      onChangeStatus(row, status) {
        let key = _(this.accessKeys).filter({'app_key': row.app_key}).first()
        if (key) {
          key.enable = status
        }
      },
      onCreate() {
        let self = this
        self.$http.post('/api/access-key/')
          .then(({data}) => {
            if (data.create_date) {
              data.create_date = data.create_date * 1000
            }
            if (data.expires) {
              data.expires = data.expires * 1000
            }
            self.accessKeys.push(data)
          })
          .catch(err => {
            let msg = '服务器异常'
            if (err.response) {
              msg = err.response.data.message
            }
            self.$message.error(msg)
          })
      },
      onUpdate(row) {
        let self = this
        let item = _.clone(row)
        if (item.expires) {
          item.expires = item.expires / 1000
        }
        self.$http.post(`/api/access-key/${row.app_key}`, item)
          .then(() => {
            self.$message.success('更新成功')
          })
          .catch(err => {
            let msg = '服务器异常'
            if (err.response) {
              msg = err.response.data.message
            }
            self.$message.error(msg)
          })
      },
      onDelete(row) {
        let self = this
        this.$confirm('AccessKey删除之后将不能恢复，是否继续', 'Warning', {
          confirmButtonText: '继续',
          cancelButtonText: '取消',
          type: 'warning'
        })
          .then(() => {
            self.$http.delete(`/api/access-key/${row.app_key}`)
              .then(() => {
                self.$message.success('删除成功')
                self.loadAccessKeys()
              })
              .catch(err => {
                let msg = '服务器异常'
                if (err.response) {
                  msg = err.response.data.message
                }
                self.$message.error(msg)
              })
          })
          .catch(() => {
          })

      },
      loadAccessKeys() {
        let self = this
        self.$http.get('/api/access-key/')
          .then(({data}) => {
            _.each(data, item => {
              if (item.create_date) {
                item.create_date = item.create_date * 1000
              }
              if (item.expires) {
                item.expires = item.expires * 1000
              }
            })
            self.accessKeys = data
          })
          .catch(err => {
            let msg = '服务器异常'
            if (err.response) {
              msg = err.response.data.message
            }
            self.$message.error(msg)
          })
      },
      loadBuckets() {
        const self = this
        this.$http.get('/api/buckets')
          .then(({data}) => {
            _.forEach(data, i => {
              i.scope = []
            })
            self.buckets = data
          })
          .catch(error => {
            self.$message(error.response.data.message)
          })
      },
    },
    mounted() {
      this.loadAccessKeys()
      this.loadBuckets()
    }
  }
</script>