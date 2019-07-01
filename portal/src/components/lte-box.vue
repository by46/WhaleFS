<template>
    <div class="box box-primary lte-box">
        <div class="box-header">
            <h3 class="box-title">{{title}}</h3>
            <div class="box-tools pull-right">
                <el-row type="flex" justify="end">
                    <el-row type="flex" justify="end">
                        <slot name="tools"></slot>
                    </el-row>
                    <button
                            v-if="collapsible"
                            class="btn btn-box-tool"
                            data-widget="collapse"><i class="fa fa-minus"></i></button>
                </el-row>
            </div>
        </div>
        <div class="box-body"
             :class="[
         isSearchPanel ? 'is-search-panel': ''
         ]">
            <slot></slot>
        </div>
    </div>
</template>

<script>
  export default {
    name: 'lte-box',
    props: {
      title: {
        type: String,
        default: '默认标题'
      },
      collapsible: {
        type: Boolean,
        default: true
      },
      isSearchPanel: {
        type: Boolean,
        default: false
      }
    }
  }
</script>

<style scoped lang="less">
    .lte-box {

        div.box-tools {
            .el-row {
                margin-right: 0;
                margin-bottom: 0px;
            }
        }

        /deep/ div.el-date-editor {
            width: 100%;
        }

        /deep/ div.el-select {
            width: 100%;
        }

        /deep/ div.el-input-number {
            width: 100%;
        }

        div.is-search-panel {
            /deep/ div.el-form-item {
                margin-bottom: 0px;
            }
        }
    }
</style>
