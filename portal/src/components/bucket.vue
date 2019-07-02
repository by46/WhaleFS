<template>
    <div>
        <el-tabs v-model="activeName">
            <el-tab-pane label="设置" name="settings">
                <el-form ref="form" :model="entity" label-width="140px" label-suffix=":"
                         :rules="rules">
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
                            <el-form-item label="存储类型" prop="collection">
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
                            <el-form-item label="备份策略" prop="replication">
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
                            <el-form-item label="别名" prop="basis.alia">
                                <el-input placeholder="别名"
                                          v-model="entity.basis.alia">
                                </el-input>
                            </el-form-item>
                        </el-col>

                        <el-col :md="16">
                            <el-form-item label="Bucket说明" prop="memo">
                                <el-input placeholder="Bucket说明"
                                          v-model="entity.memo">
                                </el-input>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10">
                        <el-col :md="8">
                            <el-form-item label="缓存过期(单位:秒)" prop="basis.expires">
                                <el-input-number v-model="entity.basis.expires"
                                                 style="width:100%;"
                                                 :step="60*60" :min="0"
                                                 :max="60*60*24*360*10"></el-input-number>
                            </el-form-item>
                        </el-col>
                        <el-col :md="8">
                            <el-form-item label="图片阈值(单位:像素)" prop="basis.prepare_thumbnail_min_width">
                                <el-input-number placeholder="图片预处理宽度阈值"
                                                 style="width:100%;"
                                                 v-model="entity.basis.prepare_thumbnail_min_width"
                                                 :step="100" :min="0"
                                                 :max="2000"></el-input-number>
                            </el-form-item>
                        </el-col>
                    </el-row>

                    <el-divider content-position="left">限制策略</el-divider>
                    <el-row :gutter="10">
                        <el-col :md="8">
                            <el-form-item label="文件最小值(单位:字节)" prop="limit.min_size">
                                <el-input-number placeholder="文件最小值"
                                                 style="width:100%;"
                                                 v-model="entity.limit.min_size"
                                                 :step="100" :min="0"></el-input-number>

                            </el-form-item>

                        </el-col>
                        <el-col :md="8">
                            <el-form-item label="文件最大值(单位:字节)" prop="limit.max_size">
                                <el-input-number placeholder="文件最大值"
                                                 style="width:100%;"
                                                 v-model="entity.limit.max_size"
                                                 :step="100" :min="0"></el-input-number>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10">
                        <el-col :md="8">
                            <el-form-item label="图片宽度" prop="limit.width">
                                <el-input-number placeholder="图片宽度"
                                                 style="width:100%;"
                                                 v-model="entity.limit.width"
                                                 :step="100" :min="0"></el-input-number>
                            </el-form-item>
                        </el-col>
                        <el-col :md="8">
                            <el-form-item label="图片高度" prop="limit.height">
                                <el-input-number placeholder="图片高度"
                                                 style="width:100%;"
                                                 v-model="entity.limit.height"
                                                 :step="100" :min="0"></el-input-number>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10">
                        <el-form-item label="Mime" prop="limit.mime_types">
                            <el-select v-model="entity.limit.mime_types" multiple placeholder="Select"
                                       style="width:100%;">
                                <el-option
                                        v-for="item in mimes"
                                        :key="item"
                                        :label="item"
                                        :value="item">
                                </el-option>
                            </el-select>
                        </el-form-item>
                    </el-row>

                    <el-divider content-position="left">图片变换</el-divider>
                    <el-row :gutter="10">
                        <el-form-item prop="sizes" label-width="0">
                            <el-table :data="entity.sizes" stripe class="bucket-sizes"
                                      style="width: 100%">
                                <el-table-column
                                        label="名称"
                                        width="140">
                                    <template slot-scope="{row, $index}">
                                        <el-form-item :prop="'sizes.'+$index+'.name'"
                                                      label-width="0"
                                                      :rules="rules.size_name">
                                            <lte-error-tip>
                                                <el-input v-model="row.name"></el-input>
                                            </lte-error-tip>
                                        </el-form-item>
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
                                        label="高度"
                                        width="180">
                                    <template slot-scope="{row}">
                                        <el-input-number v-model="row.height"
                                                         :step="10"
                                                         :min="0"
                                                         :max="2000"></el-input-number>
                                    </template>
                                </el-table-column>
                                <el-table-column label="缩放模式">
                                    <template slot-scope="{row}">
                                        <el-select v-model="row.mode">
                                            <el-option
                                                    v-for="item in modes"
                                                    :key="item.value"
                                                    :label="item.label"
                                                    :value="item.value">
                                            </el-option>
                                        </el-select>
                                    </template>
                                </el-table-column>
                                <el-table-column label="操作" width="200px">
                                    <template slot="header">
                                        <el-button type="primary" @click="onSizeAdd">新增</el-button>
                                    </template>
                                    <template slot-scope="{$index}">
                                        <el-button type="text" @click="onSizeDelete($index)">删除</el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                        </el-form-item>
                    </el-row>

                    <el-divider content-position="left">水印</el-divider>
                    <el-row :gutter="10">
                        <el-form-item prop="overlays" label-width="0">
                            <el-table :data="entity.overlays"
                                      stripe
                                      class="bucket-sizes"
                                      style="width: 100%">
                                <el-table-column
                                        label="名称"
                                        width="150">
                                    <template slot-scope="{row, $index}">
                                        <el-form-item :prop="'overlays.'+$index+'.name'"
                                                      label-width="0"
                                                      :rules="rules.overlay_name">
                                            <lte-error-tip>
                                                <el-input v-model="row.name"></el-input>
                                            </lte-error-tip>
                                        </el-form-item>

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
                                        <el-upload
                                                class="avatar-uploader"
                                                action="string"
                                                :http-request="onUploadImg(row)"
                                                :limit="1"
                                                :show-file-list="false"
                                                :on-success="handleAvatarSuccess(row)"
                                                :before-upload="beforeAvatarUpload">
                                            <img v-if="imageUrl(row.image)"
                                                 :src="imageUrl(row.image)"
                                                 class="avatar">
                                            <i v-else class="el-icon-plus avatar-uploader-icon"></i>
                                        </el-upload>
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
                                    <template slot="header">
                                        <el-button type="primary" @click="onOverlayAdd">新增</el-button>
                                    </template>
                                    <template slot-scope="{$index}">
                                        <el-button type="text" @click="onOverlayDelete($index)">删除</el-button>
                                    </template>
                                </el-table-column>
                            </el-table>
                        </el-form-item>
                    </el-row>
                    <el-row :gutter="10" style="margin-top: 20px;">
                        <el-form-item>
                            <el-button type="primary" @click="onSave">保存</el-button>
                            <el-button @click="onReturn">返回</el-button>
                        </el-form-item>
                    </el-row>
                </el-form>
            </el-tab-pane>
            <el-tab-pane label="上传/下载" name="upload">
                <el-row :gutter="10" style="margin-bottom: 20px;">
                    <el-alert type="warning"
                              :closable="false">
                        <template slot="title">
                            <div style="font-size: 12px; line-height: 24px;">
                                Bucket创建之后不能立即上传文件， 需要等待几秒同步时间<br/>
                                只能针对已经存在的Bucket进行上传
                            </div>
                        </template>
                    </el-alert>
                </el-row>
                <el-form :model="upload" label-width="140px" label-suffix=":" :disabled="!isModify">
                    <el-row :gutter="10">
                        <el-col :md="16">
                            <el-form-item prop="key" label="自定义Key">
                                <el-input v-model="upload.key" placeholder="Object Key名称，不包含bucket名称"></el-input>
                            </el-form-item>
                        </el-col>
                        <el-col :md="8">
                            <el-form-item prop="override" label-width="0">
                                <el-checkbox v-model="upload.override">是否允许覆盖</el-checkbox>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10">
                        <el-col :md="8">
                            <el-form-item label="图片尺寸">
                                <el-select v-model="upload.size">
                                    <el-option
                                            v-for="item in entity.sizes"
                                            :key="item.name"
                                            :label="item.name"
                                            :value="item.name">
                                    </el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                        <el-col :md="8">
                            <el-form-item label="水印">
                                <el-select v-model="upload.overlay">
                                    <el-option
                                            v-for="item in entity.overlays"
                                            :key="item.name"
                                            :label="item.name"
                                            :value="item.name">
                                    </el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10">
                        <el-row :md="16">
                            <el-form-item>
                                <el-upload
                                        class="upload-demo"
                                        action="string"
                                        :http-request="onUploadFile"
                                        :file-list="fileList"
                                        list-type="picture">
                                    <el-button size="small" type="primary">点击上传</el-button>
                                </el-upload>
                            </el-form-item>
                        </el-row>
                    </el-row>
                </el-form>

            </el-tab-pane>
        </el-tabs>

    </div>
</template>

<script>
  import _ from 'lodash'
  import uuidv4 from 'uuid/v4'
  import LteErrorTip from './lte-error-tip'
  import bus from '@/utils/bus'

  export default {
    name: 'bucket',
    components: {LteErrorTip},
    data() {
      let duplicateValidator = (rule, value, callback) => {
        let msg = _(value).countBy('name')
        .toPairs()
        .filter(item => item[1] > 1)
        .map(item => item[0])
        .join(',')
        if (msg) {
          callback(new Error(`存在重复的item: ${msg}`))
        } else {
          callback()
        }
      }
      return {
        fileList: [],
        activeName: 'settings',
        rules: {
          name: [
            {required: true, message: '请输入Bucket名称', trigger: 'blur'},
            {pattern: /^[a-z0-9-_]+$/, message: '请输入小写字母，数字，-', trigger: 'blur'}
          ],
          size_name: [
            {required: true, message: '请输入Resize名称', trigger: 'blur'},
            {pattern: /^[a-z0-9]+$/, message: '请输入小写字母，数字', trigger: 'blur'}
          ],
          overlay_name: [
            {required: true, message: '请输入水印名称', trigger: 'blur'},
            {pattern: /^[a-z0-9]+$/, message: '请输入小写字母，数字', trigger: 'blur'}
          ],
          sizes: [
            {validator: duplicateValidator, trigger: 'change'}
          ],
          overlays: [
            {validator: duplicateValidator, trigger: 'change'}
          ]
        },
        isModify: false,
        version: '',
        mimes: [],
        options: [],
        replications: [
          {label: '无备份', value: '000'},
          {label: '不同数据中心备份', value: '100'}],
        collections: [
          {label: '普通', value: 'general'},
          {label: '临时', value: 'tmp'},
          {label: '商品图片', value: 'product'},
          {label: '交易', value: 'trade'}
        ],
        modes: [
          {label: '等比缩放', value: 'fit'},
          {label: '填充白边', value: 'padding'},
          {label: '拉伸', value: 'stretch'},
          {label: '等比裁剪', value: 'thumbnail'}
        ],
        positions: [
          {label: '左上角', value: 'TopLeft'},
          {label: '左下角', value: 'BottomLeft'},
          {label: '右上角', value: 'TopRight'},
          {label: '右下角', value: 'BottomRight'}],
        entity: {
          name: '',
          type: 'bucket',
          memo: '',
          basis: {
            alias: '',
            collection: 'general',
            replication: '100',
            expires: 20,
            prepare_thumbnail_min_width: 1024,
            prepare_thumbnail: ''
          },
          limit: {
            width: undefined,
            height: undefined,
            min_size: undefined,
            max_size: undefined,
            mime_types: []
          },
          sizes: [],
          overlays: []
        },
        upload: {
          size: '',
          overlay: '',
          override: false,
          image: ''
        }
      }
    },
    computed: {
      imageUrl() {
        return image => {
          if (!image) {
            return ''
          }

          let url = bus.get('dfsHost')
          return `${url}/home/overlay/${image}`
        }
      },
      fileUrl() {
        if (!this.upload.image) {
          return ''
        }

        let url = bus.get('dfsHost')
        return `${url}/${this.upload.image}`
      }
    },
    methods: {
      onReturn() {
        this.$router.push({name: 'buckets'})
      },
      onSave() {
        let self = this
        self.$refs.form.validate(valid => {
          if (!valid) {
            self.$message.error('表单验证失败，请修改后在提交')
            return
          }
          if (self.isModify) {
            let body = {
              id: this.entity.name,
              version: self.version,
              basis: this.entity
            }
            self.$http.put(`/api/buckets`, body)
            .then(() => {
              self.$message.success('修改成功')
              self.load(`system.bucket.${self.entity.name}`)
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
              self.load(`system.bucket.${self.entity.name}`)
            })
            .catch(err => {
              let msg = '新增失败'
              if (err.response) {
                msg = err.response.data.message
              }
              self.$message.error(msg)
            })
          }
        })
      },
      onSizeAdd() {
        this.entity.sizes.push({name: '', width: 400, height: 300, mode: 'fit'})
      },
      onSizeDelete(index) {
        this.entity.sizes.splice(index, 1)
      },
      onOverlayAdd() {
        this.entity.overlays.push({
          'name': '',
          'default': false,
          'position': 'TopLeft',
          'image': 'watermark.png',
          'opacity': 0.8
        })
      },
      onOverlayDelete(index) {
        this.entity.overlays.splice(index, 1)
      },
      handleAvatarSuccess(row) {
        return (res, file) => {
          this.clearFiles()
          row.imageUrl = URL.createObjectURL(file.raw)
        }
      },
      beforeAvatarUpload(file) {
        const isJPG = file.type === 'image/jpeg';
        const isLt2M = file.size / 1024 / 1024 < 2;

        if (!isJPG) {
          this.$message.error('Avatar picture must be JPG format!');
        }
        if (!isLt2M) {
          this.$message.error('Avatar picture size can not exceed 2MB!');
        }
        return isJPG && isLt2M;
      },
      load(id) {
        let name = id || this.$route.query['id']
        let self = this
        self.$http.get('/api/mimetypes')
        .then(({data}) => {
          self.mimes = _.uniq(data.sort())
        })
        if (name) {
          self.$http.get(`/api/buckets/${name}`)
          .then(response => {
            self.version = response.data.version
            let entity = response.data.basis
            entity.overlays = entity.overlays || []
            entity.sizes = entity.sizes || []
            entity.basis = entity.basis || {}
            entity.basis.expires = entity.basis.expires || undefined
            entity.basis.prepare_thumbnail_min_width = entity.basis.prepare_thumbnail_min_width || undefined
            entity.limit = entity.limit || {}
            entity.limit.width = entity.limit.width || undefined
            entity.limit.height = entity.limit.height || undefined
            entity.limit.min_size = entity.limit.min_size || undefined
            entity.limit.max_size = entity.limit.max_size || undefined

            self.entity = entity
            self.isModify = true
          })
          .catch(() => {
            this.$message.error('获取Bucket信息失败')
          })
        }
      },
      onUploadImg(row) {
        let self = this
        return (item) => {
          let extension = item.file.name.split('.').pop();
          let formData = new FormData()
          let filename = uuidv4()
          let url = bus.get('dfsHost')
          formData.append('file', item.file)
          formData.append('key', `/home/overlay/${filename}.${extension}`)
          this.$http.post(url, formData)
          .then(response => {
            row.image = response.data.title
          })
          .catch(err => {
            let msg = '修改失败'
            if (err.response) {
              msg = err.response.data.message
            }
            self.$message.error(msg)
          })
        }
      },
      onUploadFile(item) {
        let self = this
        let extension = item.file.name.split('.').pop();
        let formData = new FormData()
        let filename = uuidv4()
        let url = bus.get('dfsHost')
        let bucket = self.entity.name
        let key = _.trim(self.upload.key)
        if (key) {
          key = _.trim(key, '/')
          key = `/${self.entity.name}/${key}`
        } else {
          key = `/${bucket}/${filename}.${extension}`
        }
        formData.append('file', item.file)
        formData.append('key', key)
        formData.append('override', self.upload.override)
        this.$http.post(url, formData)
        .then(({data}) => {
          let filename = data.url
          let obj = undefined
          let size = self.upload.size || 'Original'
          if (_.indexOf(filename, '/') < 0) {
            obj = {name: item.file.name, url: `${url}/pdt/${size}/${filename}`}
          } else {
            let segments = _.split(filename, '/', 2)
            if (self.upload.size) {
              filename = _.join([segments[0], self.upload.size, segments[1]], '/')
            }
            obj = {name: item.file.name, url: `${url}/${filename}`}
          }
          self.fileList.push(obj)
        })
        .catch(err => {
          let msg = '上传文件失败'
          if (err.response) {
            msg = err.response.data.message
          }
          self.$message.error(msg)
        })
      }
    },
    mounted() {
      this.load()
    }
  }
</script>

<style scoped lang="less">
    @height: 26px;
    @container-height: 28px;

    @file-height: 400px;
    @file-container-height: 400px;

    .avatar-uploader {
        height: @container-height;

        .el-upload {
            border: 1px dashed #d9d9d9;
            border-radius: 6px;
            cursor: pointer;
            position: relative;
            overflow: hidden;
        }

        .el-upload:hover {
            border-color: #409EFF;
        }
        .avatar {
            width: 50px;
            height: @container-height;
            display: block;
        }
        .avatar-uploader-icon {
            font-size: @height;
            color: #8c939d;
            width: 50px;
            height: @height;
            line-height: @height;
            text-align: center;
        }
    }

    .file-uploader {
        height: @file-container-height;

        .el-upload {
            border: 1px dashed #d9d9d9;
            border-radius: 6px;
            cursor: pointer;
            position: relative;
            overflow: hidden;
        }

        .el-upload:hover {
            border-color: #409EFF;
        }
        .avatar {
            width: 50px;
            height: @file-container-height;
            display: block;
        }
        .avatar-uploader-icon {
            font-size: @file-height;
            color: #8c939d;
            width: 400px;
            height: @file-height;
            line-height: @file-height;
            text-align: center;
        }
    }

    .bucket-sizes {
        /deep/ .el-form-item {
            margin-bottom: 0;
        }

        /deep/ div.is-required-table-header::before {
            content: '*';
            color: #f56c6c;
            margin-right: 4px;

        }
    }
</style>
