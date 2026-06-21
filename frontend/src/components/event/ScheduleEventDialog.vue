<template>
  <v-dialog v-model="show" max-width="480" content-class="tw-m-0">
    <v-card>
      <v-card-title class="tw-text-lg tw-font-medium">
        Schedule this event
      </v-card-title>
      <v-card-text>
        <div class="tw-mb-4 tw-text-sm tw-text-dark-gray">
          Pick the final date and time for this event. Everyone who opens the
          event will see it.
        </div>

        <!-- Start date + time -->
        <div class="tw-mb-1 tw-text-sm tw-text-black">Starts</div>
        <div class="tw-flex tw-items-center tw-gap-2">
          <v-text-field
            v-model="scheduledDate"
            type="date"
            label="Date"
            prepend-inner-icon="mdi-calendar"
            dense
            hide-details
            solo
          />
          <v-text-field
            v-model="scheduledTime"
            type="time"
            label="Time"
            prepend-inner-icon="mdi-clock-outline"
            dense
            hide-details
            solo
            :disabled="!scheduledDate"
          />
        </div>

        <!-- End date + time -->
        <div class="tw-mb-1 tw-mt-4 tw-text-sm tw-text-black">Ends</div>
        <div class="tw-flex tw-items-center tw-gap-2">
          <v-text-field
            v-model="endDateInput"
            type="date"
            label="Date"
            prepend-inner-icon="mdi-calendar"
            dense
            hide-details
            solo
            :disabled="!scheduledDate"
          />
          <v-text-field
            v-model="endTimeInput"
            type="time"
            label="Time"
            prepend-inner-icon="mdi-clock-outline"
            dense
            hide-details
            solo
            :disabled="!scheduledDate"
          />
        </div>

        <!-- Google Meet link -->
        <div class="tw-mb-1 tw-mt-4 tw-text-sm tw-text-black">
          Google Meet link (optional)
        </div>
        <v-text-field
          v-model="meetingLink"
          placeholder="https://meet.google.com/..."
          prepend-inner-icon="mdi-video"
          dense
          hide-details="auto"
          solo
          :error-messages="linkError"
          @input="linkError = ''"
        />

        <div v-if="errorMessage" class="tw-mt-3 tw-text-sm tw-text-red">
          {{ errorMessage }}
        </div>
      </v-card-text>
      <v-card-actions class="tw-px-4 tw-pb-4">
        <v-btn
          v-if="hasExistingSchedule"
          text
          class="tw-text-red"
          :loading="clearing"
          :disabled="saving"
          @click="clearSchedule"
        >
          Clear schedule
        </v-btn>
        <v-spacer />
        <v-btn text :disabled="saving || clearing" @click="show = false">
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          class="tw-bg-green tw-text-white"
          :loading="saving"
          :disabled="!scheduledDate || !endDateInput || clearing"
          @click="save"
        >
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { post } from "@/utils"

export default {
  name: "ScheduleEventDialog",

  props: {
    value: { type: Boolean, default: false },
    event: { type: Object, required: true },
    // Optional {startDate, endDate} (ISO) to prefill from the grid "schedule event" flow.
    prefill: { type: Object, default: null },
  },

  data: () => ({
    scheduledDate: null,
    scheduledTime: "12:00",
    endDateInput: null,
    endTimeInput: "13:00",
    meetingLink: "",
    saving: false,
    clearing: false,
    errorMessage: "",
    linkError: "",
  }),

  computed: {
    show: {
      get() {
        return this.value
      },
      set(val) {
        this.$emit("input", val)
      },
    },
    hasExistingSchedule() {
      return !!this.event?.scheduledEvent?.startDate
    },
  },

  watch: {
    value(val) {
      if (val) this.populateForm()
    },
  },

  methods: {
    /** Initialize the form fields. A grid-flow `prefill` date wins; otherwise fall back to the
     *  event's existing schedule. */
    populateForm() {
      this.errorMessage = ""
      this.linkError = ""
      this.meetingLink = this.event?.meetingLink || ""

      const pad = (n) => String(n).padStart(2, "0")
      const toFields = (val) => {
        if (!val) return null
        const d = new Date(val)
        if (isNaN(d.getTime())) return null
        return {
          date: `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`,
          time: `${pad(d.getHours())}:${pad(d.getMinutes())}`,
        }
      }

      // Start: prefer the date/time passed from the grid "schedule event" flow.
      const start = toFields(
        this.prefill?.startDate || this.event?.scheduledEvent?.startDate
      )
      if (start) {
        this.scheduledDate = start.date
        this.scheduledTime = start.time
      } else {
        this.scheduledDate = null
        this.scheduledTime = "12:00"
      }

      // End: prefer explicit end (grid/existing), else start + event duration, else start + 1h.
      let end = toFields(
        this.prefill?.endDate || this.event?.scheduledEvent?.endDate
      )
      if (!end && start) {
        const startD = new Date(`${start.date}T${start.time}`)
        const duration = Number(this.event?.duration)
        const mins =
          duration && !isNaN(duration) && duration > 0
            ? Math.round(duration * 60)
            : 60
        end = toFields(new Date(startD.getTime() + mins * 60000).toISOString())
      }
      if (end) {
        this.endDateInput = end.date
        this.endTimeInput = end.time
      } else {
        this.endDateInput = null
        this.endTimeInput = "13:00"
      }
    },

    /** Combine the start date + time inputs into a local Date, or null when incomplete */
    buildStartDate() {
      if (!this.scheduledDate) return null
      const time = this.scheduledTime || "12:00"
      const d = new Date(`${this.scheduledDate}T${time}`)
      return isNaN(d.getTime()) ? null : d
    },

    /** Combine the end date + time inputs into a local Date, or null when incomplete */
    buildEndDate() {
      if (!this.endDateInput) return null
      const time = this.endTimeInput || "00:00"
      const d = new Date(`${this.endDateInput}T${time}`)
      return isNaN(d.getTime()) ? null : d
    },

    async save() {
      const start = this.buildStartDate()
      if (!start) {
        this.errorMessage = "Please pick a valid start date and time."
        return
      }

      const end = this.buildEndDate()
      if (!end) {
        this.errorMessage = "Please pick a valid end date and time."
        return
      }
      if (end.getTime() <= start.getTime()) {
        this.errorMessage = "The end must be after the start."
        return
      }
      const endDate = end.toISOString()

      this.errorMessage = ""
      this.linkError = ""
      this.saving = true
      try {
        const eventId = this.event.shortId ?? this.event._id
        await post(`/events/${eventId}/schedule`, {
          startDate: start.toISOString(),
          endDate,
          meetingLink: this.meetingLink || "",
        })

        // Optimistically update the local event so the banner appears immediately
        this.$emit("scheduled", {
          summary: this.event?.name || "",
          startDate: start.toISOString(),
          endDate,
          meetingLink: this.meetingLink || "",
        })
        this.show = false
      } catch (err) {
        if (
          err?.status === 400 &&
          err?.parsed?.error === "invalid-meeting-link"
        ) {
          this.linkError = "Please enter a valid http(s) link."
        } else if (
          err?.status === 403 &&
          err?.parsed?.error === "user-not-event-owner"
        ) {
          this.errorMessage = "Only the event owner can schedule this event."
        } else {
          this.errorMessage = "Something went wrong. Please try again."
        }
      } finally {
        this.saving = false
      }
    },

    async clearSchedule() {
      this.errorMessage = ""
      this.linkError = ""
      this.clearing = true
      try {
        const eventId = this.event.shortId ?? this.event._id
        await post(`/events/${eventId}/schedule`, { clear: true })
        this.$emit("cleared")
        this.show = false
      } catch (err) {
        if (
          err?.status === 403 &&
          err?.parsed?.error === "user-not-event-owner"
        ) {
          this.errorMessage = "Only the event owner can schedule this event."
        } else {
          this.errorMessage = "Something went wrong. Please try again."
        }
      } finally {
        this.clearing = false
      }
    },
  },
}
</script>
