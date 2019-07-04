import Vue from 'vue'
import moment from 'moment'
import numeral from 'numeral'
import _ from 'lodash'

const fmtDate = (value, fmt) => {
  if (_.isUndefined(value)) {
    return ''
  }
  let dt = moment(value)
  if (!dt.isValid()) {
    return ''
  }
  return dt.format(fmt)
}
const fmtNumber = (value, fmt) => {
  if (!_.isNumber(value)) {
    return ''
  }
  return numeral(value).format(fmt)
}

Vue.filter('date-format', (value, fmt) => {
  return fmtDate(value, fmt)
})

Vue.filter('default', (value, defaultValue) => {
  return value || defaultValue
})

Vue.filter('lte-date-format', (value, fmt) => {
  if (!value) {
    return ''
  }
  return moment(value).format(fmt)
})

Vue.filter('lte-default', (value, defaultValue) => {
  return value || defaultValue
})

Vue.filter('lte-currency', value => {
  return fmtNumber(value, '0.00')
})

Vue.filter('lte-percentage', value => {
  return fmtNumber(value, '0%')
})

Vue.filter('lte-numeral', (value, fmt) => {
  return numeral(value).format(fmt)
})

Vue.filter('lte-datetime', value => {
  return fmtDate(value, 'YYYY-MM-DD HH:mm:ss')
})

Vue.filter('lte-date', (value, fmt = 'YYYY-MM-DD') => {
  return fmtDate(value, fmt)
})

Vue.filter('lte-enum', (value, enums, label = '') => {
  let status = _.find(enums, item => item.value === value)
  if (status) {
    return status.label
  }
  return label
})
