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
    return _.get(state.configuration, key)
  }
}