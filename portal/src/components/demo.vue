<template>
    <el-upload
            action="string"
            :http-request="onUploadImg"
            :limit="1">
        点击上传
    </el-upload>
</template>

<script>
  import {upload} from 'whalefs'

  export default {
    name: 'demo',
    methods: {
      onUploadImg(item) {
        const self = this
        const observable = upload(item.file, '/benjamin', {host: 'http://oss.yzw.cn.qa'})
        observable.subscribe({
          error: function (err) {
            self.$message.error('upload failed', err)
          },
          complete: function (data) {
            self.$message.success('upload success', data)
          }
        })
      }
    }
  }
</script>