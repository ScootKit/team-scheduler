<template>
  <div class="tw-flex tw-cursor-pointer tw-items-center">
    <v-icon class="tw-mr-1 tw-text-green" :small="isPhone">mdi-calendar-blank</v-icon>
    <span
      class="tw-font-medium tw-text-very-dark-gray"
      :class="isPhone ? 'tw-text-lg' : 'tw-text-xl'"
      >WannPassts?</span
    >
  </div>
</template>

<script>
import { isPhone } from "@/utils"

export default {
  name: "Logo",

  props: {
    // type kept for API compatibility with existing callers; WannPassts uses a single wordmark.
    type: {
      type: String,
      default: "timeful",
    },
  },

  computed: {
    isPhone() {
      return isPhone(this.$vuetify)
    },
  },
}
</script>
