<template>
    <div
            class="el-select"
            :class="[selectSize ? 'el-select--' + selectSize : '']"
            @click.stop="toggleMenu"
            v-clickoutside="handleClose">
        <div
                class="el-select__tags"
                v-if="multiple"
                ref="tags"
                :style="{ 'max-width': inputWidth - 32 + 'px', width: '100%' }">
      <span v-if="collapseTags && selected.length">
        <el-tag
                :closable="!selectDisabled"
                :size="collapseTagSize"
                :hit="selected[0].hitState"
                type="info"
                @close="deleteTag($event, selected[0])"
                disable-transitions>
          <span class="el-select__tags-text">{{ selected[0].currentLabel }}</span>
        </el-tag>
        <el-tag
                v-if="selected.length > 1"
                :closable="false"
                :size="collapseTagSize"
                type="info"
                disable-transitions>
          <span class="el-select__tags-text">+ {{ selected.length - 1 }}</span>
        </el-tag>
      </span>
            <transition-group @after-leave="resetInputHeight" v-if="!collapseTags">
                <el-tag
                        v-for="item in selected"
                        :key="getValueKey(item)"
                        :closable="!selectDisabled"
                        :size="collapseTagSize"
                        :hit="item.hitState"
                        type="info"
                        @close="deleteTag($event, item)"
                        disable-transitions>
                    <span class="el-select__tags-text">{{ item.currentLabel }}</span>
                </el-tag>
            </transition-group>

            <input
                    type="text"
                    class="el-select__input"
                    :class="[selectSize ? `is-${ selectSize }` : '']"
                    :disabled="selectDisabled"
                    :autocomplete="autoComplete || autocomplete"
                    @focus="handleFocus"
                    @blur="softFocus = false"
                    @click.stop
                    @keyup="managePlaceholder"
                    @keydown="resetInputState"
                    @keydown.down.prevent="navigateOptions('next')"
                    @keydown.up.prevent="navigateOptions('prev')"
                    @keydown.enter.prevent="selectOption"
                    @keydown.esc.stop.prevent="visible = false"
                    @keydown.delete="deletePrevTag"
                    @compositionstart="handleComposition"
                    @compositionupdate="handleComposition"
                    @compositionend="handleComposition"
                    v-model="query"
                    @input="debouncedQueryChange"
                    v-if="filterable"
                    :style="{ 'flex-grow': '1', width: inputLength / (inputWidth - 32) + '%', 'max-width': inputWidth - 42 + 'px' }"
                    ref="input">
        </div>
        <el-input
                ref="reference"
                v-model="selectedLabel"
                type="text"
                :placeholder="currentPlaceholder"
                :name="name"
                :id="id"
                :autocomplete="autoComplete || autocomplete"
                :size="selectSize"
                :disabled="selectDisabled"
                :readonly="readonly"
                :validate-event="false"
                :class="{ 'is-focus': visible }"
                @focus="handleFocus"
                @blur="handleBlur"
                @keyup.native="debouncedOnInputChange"
                @keydown.native.down.stop.prevent="navigateOptions('next')"
                @keydown.native.up.stop.prevent="navigateOptions('prev')"
                @keydown.native.enter.prevent="selectOption"
                @keydown.native.esc.stop.prevent="visible = false"
                @keydown.native.tab="visible = false"
                @paste.native="debouncedOnInputChange"
                @mouseenter.native="inputHovering = true"
                @mouseleave.native="inputHovering = false">
            <template slot="prefix" v-if="$slots.prefix">
                <slot name="prefix"></slot>
            </template>
            <template slot="suffix">
                <i v-show="!showClose" :class="['el-select__caret', 'el-input__icon', 'el-icon-' + iconClass]"></i>
                <i v-if="showClose" class="el-select__caret el-input__icon el-icon-circle-close"
                   @click="handleClearClick"></i>
            </template>
        </el-input>
        <transition
                name="el-zoom-in-top"
                @before-enter="handleMenuEnter"
                @after-leave="doDestroy">
            <el-select-menu
                    ref="popper"
                    :append-to-body="popperAppendToBody"
                    v-show="visible && emptyText !== false">
                <div style="padding: 3px 2px 5px 2px"
                     class="lte-select-popper">
                    <el-input v-model="content"
                              ref="searchBox"
                              size="small"
                              :validateEvent="false"
                              @keydown.native.enter.prevent="onSearch"
                              @input="debouncedOnQueryInputChange">
                        <template slot="suffix">
                            <i class="el-input__icon el-icon-search" style="cursor: pointer"
                               @click="onSearch"></i>
                        </template>
                    </el-input>
                </div>
                <el-scrollbar
                        tag="ul"
                        wrap-class="el-select-dropdown__wrap"
                        view-class="el-select-dropdown__list"
                        ref="scrollbar"
                        :class="{ 'is-empty': !allowCreate && query && filteredOptionsCount === 0 }"
                        v-show="options.length > 0 && !loading">
                    <el-option
                            :value="query"
                            created
                            v-if="showNewOption">
                    </el-option>
                    <slot></slot>
                </el-scrollbar>
                <template v-if="emptyText && (!allowCreate || loading || (allowCreate && options.length === 0 ))">
                    <slot name="empty" v-if="$slots.empty"></slot>
                    <p class="el-select-dropdown__empty" v-else>
                        {{ emptyText }}
                    </p>
                </template>
            </el-select-menu>

        </transition>
    </div>
</template>


<script>
  /**
   * 带搜索框的下来列表
   * @author : benjamin.c.yan
   * @Data    : 2019-02-01
   * @Time    : 15:56
   */
  import {debounce} from 'throttle-debounce'
  import ElSelect from 'element-ui/lib/select'
  import Focus from 'element-ui/lib/mixins/focus'

  export default {
    name: 'lte-select',
    componentName: 'ElSelect',
    extends: ElSelect,
    mixins: [Focus('searchBox')],
    data() {
      return {
        content: '',
        key: '',
        options: [],
        cachedOptions: [],
        createdLabel: null,
        createdSelected: false,
        selected: this.multiple ? [] : {},
        inputLength: 20,
        inputWidth: 0,
        initialInputHeight: 0,
        cachedPlaceHolder: '',
        optionsCount: 0,
        filteredOptionsCount: 0,
        visible: false,
        softFocus: false,
        selectedLabel: '',
        hoverIndex: -1,
        query: '',
        previousQuery: null,
        inputHovering: false,
        currentPlaceholder: '',
        menuVisibleOnFocus: false,
        isOnComposition: false,
        isSilentBlur: false
      }
    },
    props: {
      value: {
        required: true
      },
      url: {
        type: String
      },
      text: {
        type: String
      }
    },
    methods: {
      onSearch() {
        this.$emit('search', this.content)
      },
      onScrollDown() {
        this.$emit('scroll-down')
      },
      handleMenuEnter() {
        this.$nextTick(() => {
          this.focus()
          this.scrollToOption(this.selected)
        })
      }
    },
    created() {
      this.debouncedOnQueryInputChange = debounce(300, () => {
        this.onSearch()
      })
    }
  }
</script>

<style scoped lang="less">
    div.lte-select {
        ul.el-scrollbar__view {
            padding-top: 0px;
        }

        /deep/ input::-webkit-input-placeholder {
            color: #000;
        }

        /deep/ input::-moz-placeholder {
            color: #000;
        }

        /deep/ input:-moz-placeholder {
            color: #000;
        }

        /deep/ input:-ms-input-placeholder {
            color: #000;
        }
    }

    div.lte-select-popper {
        /deep/ div.el-input {
            i.el-input__validateIcon {
                display: none;
            }
        }
    }
</style>
