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

        <div
          v-if="showTzNote"
          class="tw-mt-3 tw-text-xs tw-text-dark-gray"
        >
          Times are in {{ tzLabel }}
        </div>

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
import { post, convertToUTC } from "@/utils"
import dayjs from "dayjs"
import utcPlugin from "dayjs/plugin/utc"
import timezonePlugin from "dayjs/plugin/timezone"
dayjs.extend(utcPlugin)
dayjs.extend(timezonePlugin)

export default {
  name: "ScheduleEventDialog",

  props: {
    value: { type: Boolean, default: false },
    event: { type: Object, required: true },
    // Optional {startDate, endDate} (ISO) to prefill from the grid "schedule event" flow.
    prefill: { type: Object, default: null },
    // IANA timezone the date/time inputs are interpreted in (the event timezone).
    // Falls back to the browser zone for legacy events.
    timezone: { type: String, default: null },
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
    /** Canonical reference timezone: the explicitly-passed event timezone, else
     *  the browser's IANA zone (legacy events). */
    tzName() {
      return this.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone
    },
    /** Only note the timezone when the inputs are NOT in the viewer's own zone
     *  (so a user picking in their own time never sees a foreign label). */
    showTzNote() {
      return this.tzName !== Intl.DateTimeFormat().resolvedOptions().timeZone
    },
    /** Short timezone label (e.g. "PDT") for the current reference timezone,
     *  derived from the chosen start instant so DST is reflected. */
    tzLabel() {
      try {
        const ref = this.scheduledDate
          ? new Date(`${this.scheduledDate}T${this.scheduledTime || "12:00"}`)
          : new Date()
        const parts = new Intl.DateTimeFormat("en-US", {
          timeZone: this.tzName,
          timeZoneName: "short",
        }).formatToParts(ref)
        const tzPart = parts.find((p) => p.type === "timeZoneName")
        return tzPart ? `${this.tzName} (${tzPart.value})` : this.tzName
      } catch (err) {
        return this.tzName
      }
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

      // Render a stored UTC instant into wall-clock date/time fields IN THE EVENT
      // TIMEZONE (not the browser's), so dragging a 9:00 AM event-tz slot shows
      // 09:00 here regardless of the viewer's browser timezone.
      const toFields = (val) => {
        if (!val) return null
        const d = dayjs(val).tz(this.tzName)
        if (!d.isValid()) return null
        return {
          date: d.format("YYYY-MM-DD"),
          time: d.format("HH:mm"),
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
        // Interpret the start wall-clock in the event timezone, add the duration,
        // then render the resulting instant back in the event timezone.
        const startInstant = convertToUTC(
          `${start.date} ${start.time}`,
          this.tzName
        )
        const duration = Number(this.event?.duration)
        const mins =
          duration && !isNaN(duration) && duration > 0
            ? Math.round(duration * 60)
            : 60
        end = toFields(
          new Date(startInstant.getTime() + mins * 60000).toISOString()
        )
      }
      if (end) {
        this.endDateInput = end.date
        this.endTimeInput = end.time
      } else {
        this.endDateInput = null
        this.endTimeInput = "13:00"
      }
    },

    /** Combine the start date + time inputs into a UTC Date, interpreting the
     *  inputs in the EVENT timezone, or null when incomplete/invalid. */
    buildStartDate() {
      if (!this.scheduledDate) return null
      const time = this.scheduledTime || "12:00"
      try {
        const d = convertToUTC(`${this.scheduledDate} ${time}`, this.tzName)
        return isNaN(d.getTime()) ? null : d
      } catch (err) {
        return null
      }
    },

    /** Combine the end date + time inputs into a UTC Date, interpreting the
     *  inputs in the EVENT timezone, or null when incomplete/invalid. */
    buildEndDate() {
      if (!this.endDateInput) return null
      const time = this.endTimeInput || "00:00"
      try {
        const d = convertToUTC(`${this.endDateInput} ${time}`, this.tzName)
        return isNaN(d.getTime()) ? null : d
      } catch (err) {
        return null
      }
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
