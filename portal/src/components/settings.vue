<template>
    <div>
        <el-divider content-position="left">App ID/App Secret Key</el-divider>
        <el-table :data="accessKeys" stripe>
            <el-table-column label="创建时间"
                             prop="create_date"
                             width="100px"
                             sortable>
                <template slot-scope="{row}">
                    {{row.create_date | lte-date}}
                </template>
            </el-table-column>
            <el-table-column label="AccessKey/SecretKey">
                <template slot-scope="{row}" width="300px">
                    <el-row>
                        <el-col :md="2">AK:</el-col>
                        <el-col :md="22">
                            <el-input v-model="row.app_id" :disabled="true"></el-input>
                        </el-col>
                    </el-row>
                    <el-row style="margin-top: 5px;">
                        <el-col :md="2">SK:</el-col>
                        <el-col :md="22">
                            <lte-input v-model="row.app_secret_key" :disabled="true" show-password></lte-input>
                        </el-col>
                    </el-row>
                </template>
            </el-table-column>
            <el-table-column label="过期时间" sortable prop="expires">
                <template slot-scope="{row}">
                    <el-date-picker
                            v-model="row.expires"
                            type="date"
                            placeholder="过期时间"
                            value-format="timestamp"
                            :change="onChangeExpires(row)"
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
            <el-table-column label="操作" width="100px">
                <template slot-scope="{row}">
                    <el-button @click="onChangeStatus(row, false)"
                               v-if="row.enable"
                               type="danger">禁用
                    </el-button>
                    <el-button @click="onChangeStatus(row, true)" v-else
                               type="primary">启用
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
        accessKeys: [{
          app_id: 'app_id',
          app_secret_key: 'app_secret_key',
          expires: 1515433133124,
          create_date: 1515433133124,
          enable: false,
          scope: ['bucket']
        }]
      }
    },
    methods: {
      onChangeStatus(row, status) {
        let key = _(this.accessKeys).filter({'app_id': row.app_id}).first()
        if (key) {
          key.enable = status
        }
      },
      onCreate() {
        this.accessKeys.push({
          app_id: 'app_id2',
          app_secret_key: 'app_secret_key2',
          expires: 1515433133124,
          create_date: 1515433133124,
          enable: false,
          scope: ['bucket']
        })
      },
      onChangeExpires(row) {
        console.log(row)
      }
    }
  }
</script>