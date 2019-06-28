<template>
    <div>
        <el-form ref="form" :model="entity" label-width="140px" label-suffix=":">
            <lte-box title="基本信息">
                <el-row :gutter="10">
                    <el-col :md="8">
                        <el-form-item label="Bucket名称" prop="name">
                            <el-input placeholder="Bucket名称"
                                      :disabled="isModify"
                                      v-model="entity.name"
                                      maxlength="128">
                            </el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="Bucket说明" prop="memo">
                            <el-input placeholder="Bucket说明"
                                      v-model="entity.memo">
                            </el-input>
                        </el-form-item>
                    </el-col>
                </el-row>
            </lte-box>
            <lte-box title="基本信息">
                <el-row :gutter="10">
                    <el-col :md="8">
                        <el-form-item label="别名" prop="basis.alia">
                            <el-input placeholder="别名"
                                      v-model="entity.basis.alia">
                            </el-input>
                        </el-form-item>
                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="存储类型" prop="basis.collection">
                            <el-select v-model="entity.basis.collection" placeholder="存储类型"
                                       :disabled="isModify">
                                <el-option
                                        v-for="item in collections"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value">
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="备份策略" prop="basis.replication">
                            <el-select v-model="entity.basis.replication" placeholder="备份策略"
                                       :disabled="isModify">
                                <el-option
                                        v-for="item in replications"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value">
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-row :gutter="10">
                    <el-col :md="8">
                        <el-form-item label="缓存过期(单位:秒)" prop="basis.expires">
                            <el-input-number v-model="entity.basis.expires"
                                             :step="60*60" :min="0"
                                             :max="60*60*24*360*10"></el-input-number>
                        </el-form-item>
                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="图片阈值(单位:像素)" prop="basis.prepare_thumbnail_min_width">
                            <el-input-number placeholder="图片预处理宽度阈值"
                                             v-model="entity.basis.prepare_thumbnail_min_width"
                                             :step="100" :min="0"
                                             :max="2000"></el-input-number>
                        </el-form-item>
                    </el-col>
                </el-row>
            </lte-box>
            <lte-box title="限制策略">
                <el-row :gutter="10">
                    <el-col :md="8">
                        <el-form-item label="文件最小值(单位:字节)" prop="limit.min_size">
                            <el-input-number placeholder="文件最小值"
                                             v-model="entity.limit.min_sie"
                                             :step="100" :min="0"></el-input-number>

                        </el-form-item>

                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="文件最小值(单位:字节)" prop="limit.max_size">
                            <el-input-number placeholder="文件最小值"
                                             v-model="entity.limit.max_size"
                                             :step="100" :min="0"></el-input-number>
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-row :gutter="10">
                    <el-col :md="8">
                        <el-form-item label="图片宽度" prop="limit.width">
                            <el-input-number placeholder="图片宽度"
                                             v-model="entity.limit.width"
                                             :step="100" :min="0"></el-input-number>
                        </el-form-item>
                    </el-col>
                    <el-col :md="8">
                        <el-form-item label="图片高度" prop="limit.height">
                            <el-input-number placeholder="图片高度"
                                             v-model="entity.limit.height"
                                             :step="100" :min="0"></el-input-number>
                        </el-form-item>
                    </el-col>
                </el-row>
                <el-row :gutter="10">
                    <el-form-item label="Mime" prop="limit.mime_types">
                        <el-select v-model="entity.limit.mime_types" multiple placeholder="Select" style="width:100%;">
                            <el-option
                                    v-for="item in mimes"
                                    :key="item"
                                    :label="item"
                                    :value="item">
                            </el-option>
                        </el-select>
                    </el-form-item>
                </el-row>
            </lte-box>

            <lte-box title="图片变换">
                <el-row>
                    <el-button @click="onSizeAdd">新增</el-button>
                </el-row>
                <el-row :gutter="10">
                    <el-table :data="entity.sizes" style="width: 100%">
                        <el-table-column
                                label="名称"
                                width="180">
                            <template slot-scope="{row}">
                                <el-input v-model="row.name"></el-input>
                            </template>
                        </el-table-column>
                        <el-table-column
                                label="宽度"
                                width="180">
                            <template slot-scope="{row}">
                                <el-input-number v-model="row.width"
                                                 :step="10"
                                                 :min="0"
                                                 :max="2000"></el-input-number>
                            </template>
                        </el-table-column>
                        <el-table-column
                                label="高度">
                            <template slot-scope="{row}">
                                <el-input-number v-model="row.height"
                                                 :step="10"
                                                 :min="0"
                                                 :max="2000"></el-input-number>
                            </template>
                        </el-table-column>
                        <el-table-column label="操作" width="200px">
                            <template slot-scope="{$index}">
                                <el-button type="text" @click="onSizeDelete($index)">删除</el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                </el-row>
            </lte-box>
            <lte-box title="水印">
                <el-row>
                    <el-button @click="onOverlayAdd">新增</el-button>
                </el-row>
            </lte-box>
            <el-row>
                <el-table
                        :data="entity.overlays" style="width: 100%">
                    <el-table-column
                            label="默认/名称"
                            width="150">
                        <template slot-scope="{row}">
                            <el-input v-model="row.name"></el-input>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="默认"
                            width="80">
                        <template slot-scope="{row}">
                            <el-checkbox v-model="row.default">默认</el-checkbox>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="位置">
                        <template slot-scope="{row}">
                            <el-select v-model="row.position" style="width:100%;">
                                <el-option
                                        v-for="item in positions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value">
                                </el-option>
                            </el-select>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="图片">
                        <template slot-scope="{row}">
                            <el-input v-model="row.image"></el-input>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="透明度">
                        <template slot-scope="{row}">
                            <el-input-number v-model="row.opacity" :precision="2" :step="0.1" :min="0.01"
                                             :max="1"></el-input-number>
                        </template>
                    </el-table-column>
                    <el-table-column label="操作" width="200px">
                        <template slot-scope="{$index}">
                            <el-button type="text" @clic="onOverlayDelete($index)">删除</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-row>
        </el-form>
        <el-row>
            <el-button @click="onSave">保存</el-button>
            <el-button @click="onReturn">返回</el-button>
        </el-row>
    </div>
</template>

<script>
  import LteBox from './lte-box'
  import _ from 'lodash'

  export default {
    name: 'bucket',
    components: {LteBox},
    data() {
      return {
        isModify: false,
        mimes: [],
        options: ['text/plain', 'image/jpeg'],
        replications: [
          {label: '无备份', value: '000'},
          {label: '不同数据中心备份', value: '100'}],
        collections: [
          {label: '普通', value: 'general'},
          {label: '临时', value: 'tmp'},
          {label: '商品图片', value: 'product'},
          {label: '交易', value: 'trade'}
        ],
        positions: [
          {label: '左上角', value: 'TopLeft'},
          {label: '左下角', value: 'BottomLeft'},
          {label: '右上角', value: 'TopRight'},
          {label: '右下角', value: 'BottomRight'}],
        entity: {
          name: 'benjamin',
          type: 'bucket',
          memo: 'testing bucket',
          'basis': {
            'alias': 'pdt',
            'collection': '',
            'replication': '100',
            'expires': 20,
            'prepare_thumbnail_min_width': 1024,
            'prepare_thumbnail': ''
          },
          limit: {
            'min_size': null,
            'max_size': 102400,
            'width': null,
            'height': null,
            'mime_types': ['image/png', 'image/jpeg', 'image/png']
          },
          sizes: [{name: 'p200', width: 200, height: 200}],
          overlays: [{
            'name': 'demo1',
            'default': true,
            'position': 'TopLeft',
            'image': '7,15154f3ef7',
            'opacity': 0.8
          }]
        }
      }
    },
    methods: {
      onReturn() {
        this.$router.push({name: 'buckets'})
      },
      onSave() {
        let self = this
        if (self.isModify) {
          let body = {
            id: this.entity.name,
            version: this.$route.query['version'],
            basis: this.entity
          }
          self.$http.put(`/api/buckets`, body)
          .then(() => {
            self.$message.success('修改成功')
          })
          .catch(err => {
            let msg = '修改失败'
            if (err.response) {
              msg = err.response.data.message
            }
            self.$message.error(msg)
          })
        } else {
          self.$http.post(`/api/buckets`, {id: this.entity.name, basis: this.entity})
          .then(() => {
            self.$message.success('新增成功')
            self.$router.push({name: 'bucket', query: {id: `system.bucket.${self.entity.name}`}})
          })
          .catch(err => {
            self.$message.error('新增失败', err.message)
          })
        }
      },
      onSizeAdd() {
        this.entity.sizes.push({name: '', width: 400, height: 300})
      },
      onSizeDelete(index) {
        this.entity.sizes.splice(index, 1)
      },
      onOverlayAdd() {
        this.entity.overlays.push({
          'name': 'name',
          'default': false,
          'position': 'TopLeft',
          'image': 'watermark.png',
          'opacity': 0.8
        })
      },
      onOverlayDelete(index) {
        this.entity.overlays.splice(index, 1)
      }
    },
    mounted() {
      let name = this.$route.query['id']
      let self = this
      self.$http.get('/api/mimetypes')
      .then(({data}) => {
        self.mimes = _.uniq(data.sort())
      })
      if (name) {
        self.$http.get(`/api/buckets/${name}`)
        .then(response => {
          self.entity = response.data
          self.entity.overlays = self.entity.overlays || []
          self.entity.sizes = self.entity.sizes || []
          self.basis = self.basis || {}
          self.limit = self.limit || {}
          self.isModify = true
        })
        .catch(() => {
          this.$message.error('获取Bucket信息失败')
        })
      }
    }
  }
</script>