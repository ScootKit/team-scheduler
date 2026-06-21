import Vue from "vue"
import VueWorker from "vue-worker"
import App from "./App.vue"
import router from "./router"
import store from "./store"
import vuetify from "./plugins/vuetify"
import posthogPlugin from "./plugins/posthog"
import VueMeta from "vue-meta"
// Self-hosted Material Design Icons (Vuetify mdi-* icons) — no external CDN (GDPR)
import "@mdi/font/css/materialdesignicons.css"
// Self-hosted DM Sans font — no Google Fonts request (GDPR)
import "@fontsource/dm-sans/400.css"
import "@fontsource/dm-sans/500.css"
import "@fontsource/dm-sans/700.css"
import "./index.css"

// Posthog (no-op stub; analytics stripped from WannPassts build)
Vue.use(posthogPlugin)

// Site Metadata
Vue.use(VueMeta)

// Workers
Vue.use(VueWorker)

Vue.config.productionTip = false

// Whether inputs should auto-focus on mount. Disabled on touch devices (phones/tablets) so that
// opening a page/dialog doesn't pop up the on-screen keyboard. Bind via :autofocus="$autofocusEnabled".
Vue.prototype.$autofocusEnabled = !(
  typeof window !== "undefined" &&
  window.matchMedia &&
  window.matchMedia("(pointer: coarse)").matches
)

new Vue({
  router,
  store,
  vuetify,
  render: (h) => h(App),
}).$mount("#app")
