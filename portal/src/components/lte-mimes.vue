<template>
    <lte-select :value="value" v-on="listeners" multiple placeholder="Select"
                @search="onSearch"
                style="width:100%;">
        <el-option-group
                v-for="group in groups"
                :key="group.name"
                :label="group.name">
            <el-option
                    v-for="item in group.items"
                    :key="item.value"
                    :label="item.name"
                    :value="item.value">
                <span style="float: left">{{ item.name }}</span>
                <span style="float: right; color: #8492a6; font-size: 13px">{{ item.value }}</span>
            </el-option>
        </el-option-group>
    </lte-select>
</template>

<script>
  import _ from 'lodash'
  import LteSelect from './lte-select'

  const category = ['application', 'audio', 'font', 'example', 'image', 'message', 'model', 'multipart', 'text', 'video']
  export default {
    name: 'lte-mimes',
    components: {LteSelect},
    props: {
      items: {
        default() {
          return []
        }
      },
      value: Array
    },
    data() {
      return {
        filter: ''
      }
    },
    computed: {
      groups() {
        let filterItems = this.items
        if (this.filter) {
          filterItems = _(filterItems).filter(t => {
            return _.includes(t.name, this.filter) || _.includes(t.value, this.filter)
          })
        }
        let types = _(filterItems).groupBy('value')
        .toPairs()
        .map(t => {
          return {value: t[0], name: _(t[1]).map(x => x.name).join(', ')}
        }).value()
        return _(types).groupBy(type => {
          const name = _.split(type.value, '/')[0]
          if (_.includes(category, name)) {
            return name
          }
          return 'misc'
        })
        .toPairs()
        .map(t => {
          return {name: t[0], items: t[1]}
        })
        .value()
      },
      listeners() {
        return {
          ...this.$listeners
        }
      }
    },
    methods: {
      onSearch(content) {
        this.filter = content
      }
    }
  }
</script>

<style scoped>
    /deep/ li.el-select-group__title {
        color: #F56C6C;
        font-size: 14px;
    }
</style>