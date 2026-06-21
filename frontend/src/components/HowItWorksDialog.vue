<template>
  <v-dialog v-model="dialog" max-width="800px" content-class="tw-m-0 pa-0">
    <v-card>
      <v-card-text class="tw-py-6 tw-text-very-dark-gray">
        How WannPassts works: create an event, share the link, and everyone
        marks their availability so you can find the best time to meet.
      </v-card-text>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="primary" text @click="dialog = false">Got it!</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
export default {
  name: "HowItWorksDialog",
  props: {
    value: Boolean, // v-model for dialog visibility
  },
  computed: {
    dialog: {
      get() {
        return this.value
      },
      set(val) {
        this.$emit("input", val)
      },
    },
  },
}
</script>
