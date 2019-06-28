<template>
    <div>
        <el-row>
            <el-alert>
                hello
            </el-alert>
        </el-row>
        <el-form ref="form" :model="entity" label-width="80px" size="mini">
            <el-row :gutter="10">
                <el-col :md="8">
                    <el-form-item label="Bucket名称" prop="name">
                        <el-input placeholder="Bucket名称"
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
            <el-row :gutter="10">
                <el-alert>
                    基本信息
                </el-alert>
            </el-row>
            <el-row :gutter="10">
                <el-col :md="8">
                    <el-form-item label="别名" prop="basis.alia">
                        <el-input placeholder="别名"
                                  v-model="entity.basis.alia">
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :md="8">
                    <el-form-item label="分组" prop="basis.collection">
                        <el-input placeholder="分组"
                                  v-model="entity.basis.collection">
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :md="8">
                    <el-form-item label="备份策略" prop="basis.replication">
                        <el-input placeholder="备份策略"
                                  v-model="entity.basis.replication">
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="10">
                <el-col :md="8">
                    <el-form-item label="缓存过期" prop="basis.expires">
                        <el-input placeholder="缓存过期时间"
                                  v-model="entity.basis.expires">
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :md="8">
                    <el-form-item label="图片阈值" prop="basis.prepare_thumbnail_min_width">
                        <el-input placeholder="图片预处理宽度阈值"
                                  v-model="entity.basis.prepare_thumbnail_min_width">
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-alert>
                限制策略
            </el-alert>
            <el-row :gutter="10">
                <el-col :md="8">
                    <el-form-item label="最小值" prop="limit.min_size">
                        <el-input placeholder="文件最小值"
                                  v-model="entity.limit.min_size">
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :md="8">
                    <el-form-item label="最大值" prop="limit.max_size">
                        <el-input placeholder="文件最大值"
                                  v-model="entity.limit.max_size">
                        </el-input>
                    </el-form-item>
                </el-col>
                <el-col :md="8">
                    <el-form-item label="宽度" prop="limit.width">
                        <el-input placeholder="宽度"
                                  v-model="entity.limit.width">
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="10">
                <el-col :md="8">
                    <el-form-item label="高度" prop="limit.height">
                        <el-input placeholder="高度"
                                  v-model="entity.limit.height">
                        </el-input>
                    </el-form-item>
                </el-col>
            </el-row>
            <el-row :gutter="10">
                <el-form-item label="Mime" prop="limit.mime_types">
                    <el-select v-model="entity.limit.mime_types" multiple placeholder="Select" style="width:100%;">
                        <el-option
                                v-for="item in options"
                                :key="item"
                                :label="item"
                                :value="item">
                        </el-option>
                    </el-select>
                </el-form-item>
            </el-row>
            <el-alert>
                图片套图设置
            </el-alert>
            <el-row>
                <el-button>新增</el-button>
            </el-row>
            <el-row :gutter="10">
                <el-table
                        :data="entity.sizes" style="width: 100%"
                        border>
                    <el-table-column
                            label="名称"
                            width="180">
                        <template slot-scope="{row,$index}">
                            <el-form-item :prop="'sizes.'+$index+'.name'"
                                          label-width="0px">
                                <el-input v-model="row.name"></el-input>
                            </el-form-item>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="宽度"
                            width="180">
                        <template slot-scope="{row,$index}">
                            <el-form-item :prop="'sizes.'+$index+'.width'"
                                          label-width="0px">
                                <el-input v-model="row.width"></el-input>
                            </el-form-item>
                        </template>
                    </el-table-column>
                    <el-table-column
                            label="高度">
                        <template slot-scope="{row,$index}">
                            <el-form-item :prop="'sizes.'+$index+'.height'"
                                          label-width="0px">
                                <el-input v-model="row.height"></el-input>
                            </el-form-item>
                        </template>
                    </el-table-column>
                    <el-table-column label="操作" width="200px">
                        <template slot-scope="{row,$index}">
                            <el-button type="text">删除</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-row>
            <el-alert>水印</el-alert>
            <el-row>
                <el-table
                        :data="entity.overlays" style="width: 100%">
                    <el-table-column
                            label="名称"
                            width="180">
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
                        <template slot-scope="{row}">
                            <el-button type="text">删除</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-row>
        </el-form>
        <el-row>
            <el-button @click="onSave">保存</el-button>
        </el-row>
    </div>
</template>

<script>
  export default {
    name: "bucket",
    data() {
      return {
        mode: 'add',
        options: ["text/plain", "image/jpeg"],
        positions: [
          {label: "左上角", value: "TopLeft"},
          {label: "左下角", value: "BottomLeft"},
          {label: "右上角", value: "TopRight"},
          {label: "右下角", value: "BottomRight"}],
        entity: {
          name: "benjamin",
          memo: "testing bucket",
          "basis": {
            "alias": "pdt",
            "collection": "",
            "replication": "100",
            "expires": 20,
            "prepare_thumbnail_min_width": 1024,
            "prepare_thumbnail": ""
          },
          limit: {
            "min_size": null,
            "max_size": 102400,
            "width": null,
            "height": null,
            "mime_types": ["image/png", "image/jpeg", "image/png"]
          },
          sizes: [{name: "p200", width: 200, height: 200}],
          overlays: [{
            "name": "demo1",
            "default": true,
            "position": "TopLeft",
            "image": "7,15154f3ef7",
            "opacity": 0.8
          }]
        }
      }
    },
    methods: {
      onSave() {
        this.$http.put("/api/")
      }
    },
    mounted() {
      let name = this.$route.query['name']
      let self = this
      if (!name) {
        self.$http.get(`/api/bucket/${name}`)
          .then(response => {
            self.entity = response.data()
            self.mode = 'edit'
          })
          .catch(() => {
            this.$message.error("获取Bucket信息失败")
          })
      }
    }
  }
</script>