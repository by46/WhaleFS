<template>
    <div :class="[
    type === 'textarea' ? 'el-textarea' : 'el-input',
    inputSize ? 'el-input--' + inputSize : '',
    {
      'is-disabled': inputDisabled,
      'is-exceed': inputExceed,
      'el-input-group': $slots.prepend || $slots.append,
      'el-input-group--append': $slots.append,
      'el-input-group--prepend': $slots.prepend,
      'el-input--prefix': $slots.prefix || prefixIcon,
      'el-input--suffix': $slots.suffix || suffixIcon || clearable || showPassword
    }
    ]"
         @mouseenter="hovering = true"
         @mouseleave="hovering = false"
    >
        <template v-if="type !== 'textarea'">
            <!-- 前置元素 -->
            <div class="el-input-group__prepend" v-if="$slots.prepend">
                <slot name="prepend"></slot>
            </div>
            <input
                    :tabindex="tabindex"
                    v-if="type !== 'textarea'"
                    class="el-input__inner"
                    v-bind="$attrs"
                    :type="showPassword ? (passwordVisible ? 'text': 'password') : type"
                    :disabled="inputDisabled"
                    :readonly="readonly"
                    :autocomplete="autoComplete || autocomplete"
                    ref="input"
                    @compositionstart="handleCompositionStart"
                    @compositionend="handleCompositionEnd"
                    @input="handleInput"
                    @focus="handleFocus"
                    @blur="handleBlur"
                    @change="handleChange"
                    :aria-label="label"
            >
            <!-- 前置内容 -->
            <span class="el-input__prefix" v-if="$slots.prefix || prefixIcon">
        <slot name="prefix"></slot>
        <i class="el-input__icon"
           v-if="prefixIcon"
           :class="prefixIcon">
        </i>
      </span>
            <!-- 后置内容 -->
            <span
                    class="el-input__suffix"
                    v-if="getSuffixVisible()">
        <span class="el-input__suffix-inner">
          <template v-if="!showClear || !showPwdVisible || !isWordLimitVisible">
            <slot name="suffix"></slot>
            <i class="el-input__icon"
               v-if="suffixIcon"
               :class="suffixIcon">
            </i>
          </template>
          <i v-if="showClear"
             class="el-input__icon el-icon-circle-close el-input__clear"
             @click="clear"
          ></i>
          <i v-if="showPwdVisible"
             class="el-input__icon el-icon-view el-input__clear"
             @click="handlePasswordVisible"
          ></i>
          <span v-if="isWordLimitVisible" class="el-input__count">
            <span class="el-input__count-inner">
              {{ textLength }}/{{ upperLimit }}
            </span>
          </span>
        </span>
        <i class="el-input__icon"
           v-if="validateState"
           :class="['el-input__validateIcon', validateIcon]">
        </i>
      </span>
            <!-- 后置元素 -->
            <div class="el-input-group__append" v-if="$slots.append">
                <slot name="append"></slot>
            </div>
        </template>
        <textarea
                v-else
                :tabindex="tabindex"
                class="el-textarea__inner"
                @compositionstart="handleCompositionStart"
                @compositionend="handleCompositionEnd"
                @input="handleInput"
                ref="textarea"
                v-bind="$attrs"
                :disabled="inputDisabled"
                :readonly="readonly"
                :autocomplete="autoComplete || autocomplete"
                :style="textareaStyle"
                @focus="handleFocus"
                @blur="handleBlur"
                @change="handleChange"
                :aria-label="label"
        >
    </textarea>
        <span v-if="isWordLimitVisible && type === 'textarea'"
              class="el-input__count">{{ textLength }}/{{ upperLimit }}</span>
    </div>
</template>
<script>
  import ElInput from 'element-ui/lib/input'

  export default {
    name: 'lte-input',
    extends: ElInput,
    computed: {
      showPwdVisible() {
        return this.showPassword &&
          (!!this.nativeInputValue || this.focused);
      }
    }
  }
</script>

<style scoped>

</style>