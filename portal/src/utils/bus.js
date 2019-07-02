import axios from 'axios'
import _ from 'lodash'

const state = {configuration: null}

export default {
  load() {
    if (state.configuration) {
      return Promise.resolve(true)
    }
    return axios.get('/api/configuration')
    .then(({data}) => {
      state.configuration = data
      return true
    })
    .catch(() => {
      state.configuration = {}
      return true
    })
  },
  get(key) {
    if (!state.configuration) {
      return undefined
    }
    let item = _.find(state.configuration, item => item.key === key)
    if (!item) {
      return undefined
    }
    return item.value
  }
}