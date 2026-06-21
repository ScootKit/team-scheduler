<template>
  <footer class="tw-bg-green tw-p-8 tw-text-white">
    <div class="tw-mx-auto tw-max-w-7xl">
      <div
        class="tw-flex tw-flex-col tw-gap-4 sm:tw-flex-row sm:tw-items-center sm:tw-justify-between"
      >
        <!-- Logo -->
        <div class="tw-flex tw-items-center">
          <v-icon class="tw-mr-1 tw-text-white">mdi-calendar-blank</v-icon>
          <span class="tw-text-2xl tw-font-medium">WannPassts?</span>
        </div>
        <!-- Legal -->
        <div class="tw-flex tw-flex-wrap tw-gap-x-6 tw-gap-y-1">
          <a
            v-if="privacyPolicyUrl"
            :href="privacyPolicyUrl"
            target="_blank"
            rel="noopener"
            class="tw-text-sm"
          >
            Privacy Policy
          </a>
          <a
            v-if="imprintUrl"
            :href="imprintUrl"
            target="_blank"
            rel="noopener"
            class="tw-text-sm"
          >
            Imprint
          </a>
          <a
            href="https://github.com/ScootKit/team-scheduler"
            target="_blank"
            rel="noopener"
            class="tw-text-sm"
          >
            Source code
          </a>
        </div>
      </div>
      <!-- Attribution / AGPL notice -->
      <div class="tw-mt-4 tw-text-xs tw-text-white/70">
        WannPassts is based on
        <a
          href="https://github.com/schej-it/timeful.app"
          target="_blank"
          rel="noopener"
          class="tw-underline"
          >Timeful</a
        >, an open-source project licensed under AGPL-3.0. Our modified
        <a
          href="https://github.com/ScootKit/team-scheduler"
          target="_blank"
          rel="noopener"
          class="tw-underline"
          >source code</a
        >
        is available under the same license.
      </div>
    </div>
  </footer>
</template>

<style scoped>
a {
  @apply tw-text-white/80 !important;
}
a:hover {
  @apply tw-text-white !important;
}
</style>

<script>
import { privacyPolicyUrl, imprintUrl } from "@/constants"

export default {
  name: "Footer",
  computed: {
    privacyPolicyUrl: () => privacyPolicyUrl,
    imprintUrl: () => imprintUrl,
  },
}
</script>
