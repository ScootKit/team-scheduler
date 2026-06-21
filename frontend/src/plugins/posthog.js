// PostHog analytics have been stripped from the WannPassts internal build.
//
// Upstream initialized PostHog and sent events to https://e.timeful.app (a domain
// owned by the original Timeful developers). We do not load posthog-js at all and
// expose a no-op stub so the existing `this.$posthog?.capture(...)` call sites stay
// safe without making any network calls or bundling the analytics library.
const noop = () => {}

const stub = {
  capture: noop,
  identify: noop,
  reset: noop,
  get_distinct_id: () => null,
  isFeatureEnabled: () => false,
  getFeatureFlag: () => undefined,
  onFeatureFlags: noop,
  setPersonPropertiesForFlags: noop,
}

export default {
  install(Vue) {
    Vue.prototype.$posthog = stub
  },
}
