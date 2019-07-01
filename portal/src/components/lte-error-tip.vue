<template>
    <el-tooltip popper-class="lte-error__container"
                effect="light"
                placement="right-start"
                :disabled="!content">
        <template>
            <slot></slot>
        </template>
        <template slot="content">{{content}}</template>
    </el-tooltip>
</template>

<script>
  export default {
    name: 'lte-error-tip',
    inject: ['elFormItem'],
    computed: {
      content() {
        return this.elFormItem && this.elFormItem.validateMessage
      }
    }
  }
</script>