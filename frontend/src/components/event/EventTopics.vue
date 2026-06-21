<template>
  <div class="tw-mx-4 tw-mt-8 tw-max-w-3xl sm:tw-mx-auto">
    <div
      class="tw-rounded-lg tw-border tw-border-light-gray-stroke tw-bg-white tw-p-4 sm:tw-p-5"
    >
      <div class="tw-mb-3 tw-flex tw-items-center tw-gap-2">
        <v-icon small class="tw-text-very-dark-gray"
          >mdi-message-text-outline</v-icon
        >
        <h2 class="tw-text-base tw-font-medium tw-text-black sm:tw-text-lg">
          Suggested topics
        </h2>
      </div>

      <!-- Topic list -->
      <div v-if="topics.length > 0" class="tw-space-y-2">
        <div
          v-for="topic in topics"
          :key="topic.id"
          class="tw-rounded-md tw-border tw-border-light-gray-stroke tw-bg-light-gray tw-p-3"
        >
          <div
            class="tw-whitespace-pre-wrap tw-break-words tw-text-sm tw-text-black sm:tw-text-base"
          >
            {{ topic.text }}
          </div>
          <div
            class="tw-mt-1 tw-flex tw-items-center tw-gap-1.5 tw-text-xs tw-text-dark-gray"
          >
            <span v-if="topic.name" class="tw-font-medium">{{
              topic.name
            }}</span>
            <span v-if="topic.name && topic.createdAt">·</span>
            <span v-if="topic.createdAt">{{ formatTime(topic.createdAt) }}</span>
          </div>
        </div>
      </div>
      <div v-else class="tw-text-sm tw-text-dark-gray">
        No topics suggested yet.
      </div>

      <!-- Suggest a topic -->
      <div v-if="canSuggest" class="tw-mt-4">
        <v-textarea
          ref="suggestInput"
          v-model="newTopicText"
          placeholder="Suggest a topic..."
          rows="1"
          auto-grow
          :outlined="!highlightSuggest"
          :solo="highlightSuggest"
          :class="
            highlightSuggest
              ? 'tw-ring-2 tw-ring-green tw-rounded'
              : 'tw-text-sm sm:tw-text-base'
          "
          dense
          hide-details
        ></v-textarea>
        <div class="tw-mt-2 tw-flex tw-justify-end">
          <v-btn
            class="tw-bg-green tw-text-white"
            :disabled="newTopicText.trim().length === 0"
            :loading="submitting"
            @click="submitTopic"
          >
            Suggest topic
          </v-btn>
        </div>
      </div>
      <div
        v-else-if="deadlinePassed && hasResponded"
        class="tw-mt-4 tw-text-sm tw-text-dark-gray"
      >
        Responses are closed for this event.
      </div>
    </div>
  </div>
</template>

<script>
import { mapActions } from "vuex"
import { post, formatShortTimeAgo } from "@/utils"

export default {
  name: "EventTopics",

  props: {
    event: {
      type: Object,
      required: true,
    },
    /** Whether the current user has saved their availability */
    hasResponded: {
      type: Boolean,
      default: false,
    },
    /** Whether the event's response deadline has passed */
    deadlinePassed: {
      type: Boolean,
      default: false,
    },
    /** Suggested author name to prefill the topic with */
    authorName: {
      type: String,
      default: "",
    },
  },

  data() {
    return {
      newTopicText: "",
      submitting: false,
      highlightSuggest: false,
    }
  },

  computed: {
    topics() {
      return this.event.topics ?? []
    },
    /** Only allow suggesting once responded and before the deadline */
    canSuggest() {
      return this.hasResponded && !this.deadlinePassed
    },
  },

  methods: {
    ...mapActions(["showError", "showInfo"]),

    formatTime(date) {
      return formatShortTimeAgo(date)
    },

    /** Called by the parent right after the user saves availability: scroll the
     *  suggest box into view, focus it, and briefly highlight it to nudge the
     *  user to add a topic. */
    focusSuggest() {
      if (!this.canSuggest) return
      this.$nextTick(() => {
        this.$el.scrollIntoView({ behavior: "smooth", block: "center" })
        const input = this.$refs.suggestInput
        if (input && input.focus) input.focus()
        this.highlightSuggest = true
        setTimeout(() => {
          this.highlightSuggest = false
        }, 2500)
      })
    },

    async submitTopic() {
      const text = this.newTopicText.trim()
      if (text.length === 0 || this.submitting) return

      this.submitting = true
      try {
        const topic = await post(`/events/${this.event._id}/topics`, {
          name: this.authorName ?? "",
          text,
        })

        // Append to the local event so it shows immediately
        if (!Array.isArray(this.event.topics)) {
          this.$set(this.event, "topics", [])
        }
        this.event.topics.push(topic)

        this.newTopicText = ""
      } catch (err) {
        if (err?.parsed?.error === "response-deadline-passed") {
          this.showInfo("Responses are closed for this event.")
        } else {
          console.error(err)
          this.showError("Failed to suggest topic. Please try again.")
        }
      } finally {
        this.submitting = false
      }
    },
  },
}
</script>
