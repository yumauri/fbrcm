import DefaultTheme from 'vitepress/theme'
import ContentTabs from './components/ContentTabs.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('ContentTabs', ContentTabs)
  }
}
