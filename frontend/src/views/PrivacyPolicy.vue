<template>
  <div class="tw-flex tw-h-full tw-items-center tw-justify-center tw-p-8">
    <p class="tw-text-very-dark-gray">Privacy policy not configured.</p>
  </div>
</template>

<script>
import { privacyPolicyUrl } from "@/constants"

export default {
  name: "PrivacyPolicy",

  metaInfo: {
    title: "Privacy Policy - WannPassts",
  },

  mounted() {
    if (privacyPolicyUrl) {
      window.location.replace(privacyPolicyUrl)
    }
  },
}
</script>
