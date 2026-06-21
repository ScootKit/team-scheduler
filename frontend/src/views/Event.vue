<template>
  <span>
    <div v-if="event" class="tw-mt-8 tw-h-full">
      <!-- Mark availability option dialog -->
      <MarkAvailabilityDialog
        v-model="choiceDialog"
        @allowGoogleCalendar="
          () => setAvailabilityAutomatically(calendarTypes.GOOGLE)
        "
        @setAvailabilityManually="setAvailabilityManually"
        @addedICSCalendar="addedICSCalendar"
      />

      <!-- Google sign in not supported dialog -->
      <SignInNotSupportedDialog v-model="webviewDialog" />

      <!-- Guest dialog -->
      <GuestDialog
        v-model="guestDialog"
        @submit="handleGuestDialogSubmit"
        :event="event"
        :respondents="Object.keys(event.responses)"
      />

      <!-- Join sign up slot dialog-->
      <SignUpForSlotDialog
        v-if="currSignUpBlock"
        v-model="signUpForSlotDialog"
        :signUpBlock="currSignUpBlock"
        @submit="signUpForBlock"
        :event="event"
      />

      <!-- Edit event dialog -->
      <NewDialog
        v-model="editEventDialog"
        :type="eventType"
        :event="event"
        :contactsPayload="contactsPayload"
        edit
        no-tabs
      />

      <!-- Schedule event dialog (owner only) -->
      <ScheduleEventDialog
        v-if="canEdit"
        v-model="scheduleEventDialog"
        :event="event"
        :prefill="schedulePrefill"
        :timezone="scheduleDialogTimezone"
        @scheduled="onEventScheduled"
        @cleared="onEventScheduleCleared"
      />

      <!-- Group invitation dialog -->
      <InvitationDialog
        v-if="isGroup"
        v-model="invitationDialog"
        :group="event"
        :calendarPermissionGranted="calendarPermissionGranted"
        @refreshEvent="refreshEvent"
        @setAvailabilityAutomatically="setAvailabilityAutomatically"
      ></InvitationDialog>

      <!-- Pages Not Visited dialog -->
      <v-dialog
        v-model="pagesNotVisitedDialog"
        max-width="400"
        content-class="tw-m-0"
      >
        <v-card>
          <v-card-title>Are you sure?</v-card-title>
          <v-card-text
            ><span class="tw-font-medium"
              >You're about to add your availability without filling out all
              pages of this event.</span
            >
            Click the left and right arrows at the top to switch between
            pages.</v-card-text
          >
          <v-card-actions>
            <v-spacer />
            <v-btn text @click="pagesNotVisitedDialog = false">Cancel</v-btn>
            <v-btn
              text
              color="primary"
              @click="
                () => {
                  saveChanges(true)
                  this.pagesNotVisitedDialog = false
                }
              "
              >Add anyways</v-btn
            >
          </v-card-actions>
        </v-card>
      </v-dialog>

      <div
        class="tw-mx-auto tw-mt-4 lg:tw-flex lg:tw-items-start lg:tw-justify-center lg:tw-gap-6"
      >
        <div class="tw-mx-auto tw-max-w-5xl tw-flex-1">
          <div v-if="!isSettingSpecificTimes" class="tw-mx-4">
            <!-- Title and copy link -->
            <div class="tw-flex tw-items-center tw-text-black">
              <div>
                <div
                  class="sm:mb-2 tw-flex tw-flex-wrap tw-items-center tw-gap-x-4 tw-gap-y-2"
                >
                  <div
                    class="tw-text-xl sm:tw-text-3xl"
                    :class="
                      canEdit &&
                      '-tw-mx-2 -tw-my-1 tw-cursor-pointer tw-rounded tw-px-2 tw-py-1 tw-transition-all hover:tw-bg-light-gray'
                    "
                    @click="canEdit && editEvent()"
                  >
                    {{ event.name }}
                  </div>
                  <v-chip
                    v-if="event.when2meetHref?.length > 0"
                    :href="`https://when2meet.com${event.when2meetHref}`"
                    :small="isPhone"
                    class="tw-cursor-pointer tw-select-none tw-rounded tw-bg-light-gray tw-px-2 tw-font-medium sm:tw-px-3"
                    >Imported from when2meet</v-chip
                  >
                  <template v-if="isGroup">
                    <div class="">
                      <v-chip
                        :small="isPhone"
                        class="tw-cursor-pointer tw-select-none tw-rounded tw-bg-light-gray tw-px-2 tw-font-medium sm:tw-px-3"
                        @click="helpDialog = true"
                        >Availability group</v-chip
                      >
                    </div>
                    <HelpDialog v-model="helpDialog">
                      <template v-slot:header>Availability group</template>
                      <div class="mb-4">
                        Use availability groups to see group members' weekly
                        calendar availabilities from Google Calendar. Your
                        actual calendar events are NOT visible to others.
                      </div>
                    </HelpDialog>
                  </template>
                </div>
                <div class="tw-flex tw-items-baseline tw-gap-1">
                  <div
                    class="tw-text-sm tw-font-normal tw-text-very-dark-gray sm:tw-text-base"
                  >
                    {{ dateString }}
                  </div>
                  <template v-if="canEdit">
                    <v-btn
                      id="edit-event-btn"
                      @click="editEvent"
                      class="tw-px-2 tw-text-sm tw-text-green"
                      text
                    >
                      Edit {{ isGroup ? "group" : "event" }}
                    </v-btn>
                    <v-btn
                      v-if="!isGroup && !isSignUp"
                      id="schedule-event-btn"
                      @click="openScheduleDialog"
                      class="tw-px-2 tw-text-sm tw-text-green"
                      text
                    >
                      <v-icon small class="tw-mr-1">mdi-calendar-check</v-icon>
                      Schedule this event
                    </v-btn>
                  </template>
                </div>
              </div>
              <v-spacer />
              <div class="tw-flex tw-flex-row tw-items-center tw-gap-2.5">
                <div v-if="isGroup">
                  <v-btn
                    v-if="
                      event.startOnMonday ? weekOffset != 1 : weekOffset != 0
                    "
                    :icon="isPhone"
                    text
                    class="tw-mr-1 tw-text-very-dark-gray sm:tw-mr-2.5"
                    @click="resetWeekOffset"
                  >
                    <v-icon class="sm:tw-mr-2">mdi-calendar-today</v-icon>
                    <span v-if="!isPhone">Today</span>
                  </v-btn>
                  <v-btn
                    :icon="isPhone"
                    :outlined="!isPhone"
                    class="tw-text-green"
                    @click="refreshCalendar"
                    :loading="loading"
                  >
                    <v-icon class="tw-mr-1" v-if="!isPhone">mdi-refresh</v-icon>
                    <span v-if="!isPhone" class="tw-mr-2">Refresh</span>
                    <v-icon class="tw-text-green" v-else>mdi-refresh</v-icon>
                  </v-btn>
                </div>
                <div v-else>
                  <v-btn
                    :icon="isPhone"
                    :outlined="!isPhone"
                    class="tw-text-green"
                    @click="copyLink"
                  >
                    <span v-if="!isPhone" class="tw-mr-2 tw-text-green"
                      >Copy link</span
                    >
                    <v-icon class="tw-text-green" v-if="!isPhone"
                      >mdi-content-copy</v-icon
                    >
                    <v-icon class="tw-text-green" v-else>mdi-share</v-icon>
                  </v-btn>
                </div>
                <div
                  v-if="!isPhone && !canRespond && !isGroup"
                  class="tw-flex tw-w-40 tw-items-center tw-justify-end tw-text-sm tw-font-medium tw-text-very-dark-gray"
                >
                  Responses are closed.
                </div>
                <div
                  v-else-if="!isPhone && (!isSignUp || canEdit)"
                  class="tw-flex tw-w-40"
                >
                  <template v-if="!isEditing">
                    <v-btn
                      v-if="!isGroup && !authUser && selectedGuestRespondent"
                      min-width="10.25rem"
                      class="tw-bg-green tw-text-white tw-transition-opacity"
                      :style="{ opacity: availabilityBtnOpacity }"
                      @click="editGuestAvailability"
                    >
                      {{
                        event.blindAvailabilityEnabled
                          ? "Edit availability"
                          : `Edit ${selectedGuestRespondent}'s availability`
                      }}
                    </v-btn>
                    <v-btn
                      v-else
                      width="10.25rem"
                      class="tw-text-white tw-transition-opacity"
                      :class="'tw-bg-green'"
                      :disabled="loading && !userHasResponded"
                      :style="{ opacity: availabilityBtnOpacity }"
                      @click="() => addAvailability()"
                    >
                      {{ actionButtonText }}
                    </v-btn>
                  </template>
                  <template v-else>
                    <v-btn
                      class="tw-mr-1 tw-w-20 tw-text-red"
                      @click="cancelEditing"
                      outlined
                    >
                      Cancel
                    </v-btn>
                    <v-btn
                      class="tw-w-20 tw-text-white"
                      :class="'tw-bg-green'"
                      @click="() => saveChanges()"
                    >
                      Save
                    </v-btn></template
                  >
                </div>
              </div>
            </div>

            <!-- Response deadline banner -->
            <div v-if="responseDeadline" class="tw-mt-2">
              <v-chip
                v-if="deadlinePassed"
                label
                :small="isPhone"
                class="tw-bg-light-gray tw-font-medium tw-text-red"
              >
                <v-icon small class="tw-mr-1 tw-text-red">mdi-lock</v-icon>
                Responses closed on {{ deadlineFormatted }}
              </v-chip>
              <v-chip
                v-else
                label
                :small="isPhone"
                class="tw-bg-ligher-green tw-font-medium tw-text-green"
              >
                <v-icon small class="tw-mr-1 tw-text-green"
                  >mdi-clock-outline</v-icon
                >
                Responses close {{ deadlineFormatted
                }}<template v-if="deadlineRelative">
                  ({{ deadlineRelative }})</template
                >
              </v-chip>
            </div>

            <!-- Description -->
            <EventDescription
              :event.sync="event"
              :canEdit="event.ownerId != 0 && canEdit"
            />
          </div>

          <!-- Scheduled event banner (shown to everyone once a date is set) -->
          <div
            v-if="scheduledEventStartFormatted"
            class="tw-mx-4 tw-mt-4 tw-flex tw-flex-col tw-gap-2 tw-rounded-lg tw-border tw-border-solid tw-border-green tw-bg-ligher-green tw-p-4 sm:tw-flex-row sm:tw-items-center sm:tw-justify-between"
          >
            <div class="tw-flex tw-items-center tw-gap-3">
              <v-icon class="tw-text-green" large>mdi-calendar-check</v-icon>
              <div>
                <div class="tw-text-xs tw-font-medium tw-text-green">
                  This event is scheduled for
                </div>
                <div class="tw-text-base tw-font-semibold tw-text-black">
                  📅 {{ scheduledEventStartFormatted
                  }}<span
                    v-if="scheduledRelative"
                    class="tw-ml-1 tw-font-medium tw-text-green"
                    >({{ scheduledRelative }})</span
                  >
                </div>
              </div>
            </div>
            <a
              v-if="meetingLinkUrl"
              :href="meetingLinkUrl"
              target="_blank"
              rel="noopener"
              class="tw-no-underline"
            >
              <v-btn class="tw-bg-green tw-text-white" small>
                <v-icon small class="tw-mr-1">mdi-video</v-icon>
                Join Google Meet
              </v-btn>
            </a>
          </div>

          <!-- Calendar -->

          <ScheduleOverlap
            ref="scheduleOverlap"
            :event="event"
            :ownerIsPremium="ownerIsPremium"
            :fromEditEvent="fromEditEvent"
            :loadingCalendarEvents="loading"
            :calendarEventsMap="calendarEventsMap"
            :calendarPermissionGranted="calendarPermissionGranted"
            :calendar-availabilities="calendarAvailabilities"
            :weekOffset.sync="weekOffset"
            :curGuestId="curGuestId"
            :initial-timezone="initialTimezone"
            @update:selectedTimezone="selectedTimezone = $event"
            :addingAvailabilityAsGuest="addingAvailabilityAsGuest"
            @addAvailability="addAvailability"
            @addAvailabilityAsGuest="addAvailabilityAsGuest"
            @refreshEvent="refreshEvent"
            @highlightAvailabilityBtn="highlightAvailabilityBtn"
            @deleteAvailability="deleteAvailability"
            @setCurGuestId="(id) => (curGuestId = id)"
            @signUpForBlock="initiateSignUpFlow"
            @openScheduleDialog="openScheduleDialogPrefilled"
          />
        </div>
      </div>

      <!-- Suggested topics -->
      <EventTopics
        v-if="
          topicsEnabled && !isSettingSpecificTimes && !isGroup && !isSignUp
        "
        ref="eventTopics"
        :event="event"
        :hasResponded="currentUserHasResponded"
        :deadlinePassed="deadlinePassed"
        :authorName="topicAuthorName"
      />

      <div class="tw-h-8"></div>
      <!-- Bottom bar for phones -->
      <div
        v-if="!isSettingSpecificTimes && isPhone && (!isSignUp || canEdit)"
        class="tw-fixed tw-bottom-0 tw-z-20 tw-flex tw-w-full tw-flex-col"
      >
        <div
          class="tw-flex tw-h-[4rem] tw-w-full tw-items-center tw-px-4"
          :class="`${isIOS ? 'tw-pb-2' : ''} ${
            isScheduling ? 'tw-bg-blue' : 'tw-bg-green'
          }`"
        >
          <template v-if="!isEditing && !isScheduling">
            <v-btn
              v-if="!event.daysOnly && numResponses > 0"
              text
              class="tw-text-white"
              @click="scheduleEvent"
              >Schedule</v-btn
            >
            <v-spacer />
            <template v-if="!canRespond && !isGroup">
              <span class="tw-text-sm tw-font-medium tw-text-white">
                Responses are closed.
              </span>
            </template>
            <template v-else>
              <v-btn
                v-if="!isGroup && !authUser && selectedGuestRespondent"
                class="tw-bg-white tw-text-green tw-transition-opacity"
                :style="{ opacity: availabilityBtnOpacity }"
                @click="editGuestAvailability"
              >
                {{ mobileGuestActionButtonText }}
              </v-btn>
              <v-btn
                v-else
                class="tw-bg-white tw-text-green tw-transition-opacity"
                :disabled="loading && !userHasResponded"
                :style="{ opacity: availabilityBtnOpacity }"
                @click="() => addAvailability()"
              >
                {{ mobileActionButtonText }}
              </v-btn>
            </template>
          </template>
          <template v-else-if="isEditing">
            <v-btn text class="tw-text-white" @click="cancelEditing">
              Cancel
            </v-btn>
            <v-spacer />
            <v-btn
              class="tw-bg-white tw-text-green"
              @click="() => saveChanges()"
            >
              Save
            </v-btn>
          </template>
          <template v-else-if="isScheduling">
            <v-btn text class="tw-text-white" @click="cancelScheduleEvent">
              Cancel
            </v-btn>
            <v-spacer />
            <v-btn
              :disabled="!allowScheduleEvent"
              class="tw-bg-white tw-text-blue"
              @click="confirmScheduleEvent"
            >
              Schedule
            </v-btn>
          </template>
        </div>
      </div>
    </div>
  </span>
</template>

<script>
import {
  get,
  post,
  signInGoogle,
  signInOutlook,
  isPhone,
  processEvent,
  getCalendarEventsMap,
  getDateRangeStringForEvent,
  isIOS,
  isDstObserved,
  doesDstExist,
  formatTimeUntil,
  formatInTimezone,
} from "@/utils"
import { isBetween } from "@/utils/general_utils"
import { validateEmail } from "@/utils"
import { mapActions, mapState, mapMutations, mapGetters } from "vuex"
import dayjs from "dayjs"
import utcPlugin from "dayjs/plugin/utc"
import timezonePlugin from "dayjs/plugin/timezone"
dayjs.extend(utcPlugin)
dayjs.extend(timezonePlugin)

import NewDialog from "@/components/NewDialog.vue"
import ScheduleOverlap from "@/components/schedule_overlap/ScheduleOverlap.vue"
import GuestDialog from "@/components/GuestDialog.vue"
import SignUpForSlotDialog from "@/components/sign_up_form/SignUpForSlotDialog.vue"
import {
  errors,
  authTypes,
  eventTypes,
  calendarTypes,
  dayIndexToDayString,
  allTimezones,
  guestUserId,
} from "@/constants"
import isWebview from "is-ua-webview"
import SignInNotSupportedDialog from "@/components/SignInNotSupportedDialog.vue"
import MarkAvailabilityDialog from "@/components/calendar_permission_dialogs/MarkAvailabilityDialog.vue"
import InvitationDialog from "@/components/groups/InvitationDialog.vue"
import HelpDialog from "@/components/HelpDialog.vue"
import EventDescription from "@/components/event/EventDescription.vue"
import EventTopics from "@/components/event/EventTopics.vue"
import ScheduleEventDialog from "@/components/event/ScheduleEventDialog.vue"
export default {
  name: "Event",

  props: {
    eventId: { type: String, required: true },
    fromSignIn: { type: Boolean, default: false },
    editingMode: { type: Boolean, default: false },
    linkApple: { type: Boolean, default: false },
    initialTimezone: { type: Object, default: () => ({}) },
    contactsPayload: { type: Object, default: () => ({}) },
  },

  components: {
    GuestDialog,
    SignUpForSlotDialog,
    ScheduleOverlap,
    NewDialog,
    SignInNotSupportedDialog,
    MarkAvailabilityDialog,
    InvitationDialog,
    HelpDialog,
    EventDescription,
    EventTopics,
    ScheduleEventDialog,
  },

  data: () => ({
    fromEditEvent: false,

    choiceDialog: false,
    webviewDialog: false,
    guestDialog: false,
    signUpForSlotDialog: false,
    editEventDialog: false,
    invitationDialog: false,
    pagesNotVisitedDialog: false,
    helpDialog: false,
    scheduleEventDialog: false,
    schedulePrefill: null, // {startDate, endDate} ISO when opened from the grid "schedule event" flow

    loading: true,
    calendarEventsMap: {},
    event: null,
    scheduleOverlapComponent: null,
    scheduleOverlapComponentLoaded: false,

    ownerIsPremium: false,
    ownerPremiumChecked: false,

    curGuestId: "", // Id of the current guest being edited
    // The timezone currently selected in the grid's "shown in" selector, surfaced from
    // ScheduleOverlap so page-level displays (scheduled banner, deadline) follow it.
    selectedTimezone: null,
    calendarPermissionGranted: true,
    addingAvailabilityAsGuest: false, // Whether a signed in user is current adding availability as a guest

    weekOffset: 0,

    availabilityBtnOpacity: 1,
    hasRefetchedAuthUserCalendarEvents: false,

    // Availability Groups
    calendarAvailabilities: {}, // maps userId to their calendar events

    // Sign Up Forms
    currSignUpBlock: null,

    // Ticks every 30s so the deadline countdown ("in X minutes") stays live and the
    // closed/open state flips without a manual refresh.
    nowTs: Date.now(),
    deadlineTimer: null,
  }),

  beforeMount() {},

  mounted() {
    this.deadlineTimer = setInterval(() => {
      this.nowTs = Date.now()
    }, 30000)
    // If coming from enabling contacts, show the dialog. Checks if contactsPayload is not an Observer.
    this.editEventDialog = Object.keys(this.contactsPayload).length > 0
    // If coming from signing in to link apple calendar, show the mark availability dialog
    if (this.linkApple) {
      this.choiceDialog = true
    }
  },

  computed: {
    ...mapState(["authUser", "events"]),
    ...mapGetters(["isPremiumUser"]),
    allowScheduleEvent() {
      return this.scheduleOverlapComponent?.allowScheduleEvent
    },
    calendarTypes() {
      return calendarTypes
    },
    dateString() {
      return getDateRangeStringForEvent(this.event)
    },
    isEditing() {
      return this.scheduleOverlapComponent?.editing
    },
    isScheduling() {
      return this.scheduleOverlapComponent?.scheduling
    },
    canEdit() {
      return (
        this.event.ownerId == 0 || this.authUser?._id === this.event.ownerId
      )
    },
    isPhone() {
      return isPhone(this.$vuetify)
    },
    isSpecificDates() {
      return this.event?.type === eventTypes.SPECIFIC_DATES || !this.event?.type
    },
    isWeekly() {
      return this.event?.type === eventTypes.DOW
    },
    isGroup() {
      return this.event?.type === eventTypes.GROUP
    },
    isSignUp() {
      return this.event?.isSignUpForm
    },
    eventType() {
      if (this.isGroup) return "group"
      else if (this.isSignUp) return "signup"
      else return "event"
    },
    areUnsavedChanges() {
      return this.scheduleOverlapComponent?.unsavedChanges
    },
    userHasResponded() {
      return this.authUser?._id in this.event.responses
    },
    /** Whether the current user (auth or guest) has saved their availability */
    currentUserHasResponded() {
      if (this.userHasResponded) return true
      return Boolean(this.scheduleOverlapComponent?.guestAddedAvailability)
    },
    /** Suggested author name to prefill new topics with */
    topicAuthorName() {
      if (this.authUser) {
        return `${this.authUser.firstName || ""} ${
          this.authUser.lastName || ""
        }`.trim()
      }
      return this.scheduleOverlapComponent?.guestName || ""
    },
    selectedGuestRespondent() {
      return this.scheduleOverlapComponent?.selectedGuestRespondent
    },
    showFeedbackBtn() {
      return this.isPhone
    },
    numResponses() {
      return this.scheduleOverlapComponent?.respondents.length
    },
    actionButtonText() {
      if (this.isSignUp) return "Edit slots"
      else if (this.userHasResponded || this.isGroup) return "Edit availability"
      return "Add availability"
    },
    mobileGuestActionButtonText() {
      return this.event.blindAvailabilityEnabled
        ? "Edit availability"
        : `Edit ${this.selectedGuestRespondent}'s availability`
    },
    mobileActionButtonText() {
      if (this.isSignUp) return "Edit slots"
      return this.userHasResponded ? "Edit availability" : "Add availability"
    },
    isIOS() {
      return isIOS()
    },
    isSettingSpecificTimes() {
      return (
        this.scheduleOverlapComponent?.state ===
        this.scheduleOverlapComponent?.states.SET_SPECIFIC_TIMES
      )
    },
    /** Whether respondents may suggest topics (creator opt-out; default enabled) */
    topicsEnabled() {
      return this.event?.topicsEnabled !== false
    },
    /** The event's response deadline as a Date, or null when unset */
    responseDeadline() {
      if (!this.event?.responseDeadline) return null
      const d = new Date(this.event.responseDeadline)
      return isNaN(d.getTime()) ? null : d
    },
    /** Whether the response deadline (if any) has passed */
    deadlinePassed() {
      return (
        this.responseDeadline != null &&
        this.responseDeadline.getTime() <= this.nowTs
      )
    },
    /** Canonical reference timezone for this event's scheduled/deadline displays:
     *  the persisted event timezone, else the grid's active timezone, else the
     *  browser zone (legacy events). Every viewer sees the same labeled time. */
    /** Display timezone for scheduled time + deadline: follows the grid's "shown in"
     *  selector when set, else the viewer's OWN local timezone. formatInTimezone only
     *  adds a TZ label when this differs from the browser zone, so viewing in your own
     *  zone stays label-free while picking another zone shows it explicitly. */
    eventTimezone() {
      return (
        this.selectedTimezone?.value ||
        Intl.DateTimeFormat().resolvedOptions().timeZone
      )
    },
    /** Timezone the schedule dialog interprets inputs in: when opened from the
     *  grid, match the grid's active timezone (so the dragged time lines up);
     *  otherwise the viewer's local timezone. */
    scheduleDialogTimezone() {
      return this.schedulePrefill?.timezone || this.eventTimezone
    },
    /** Human-readable deadline in the viewer's local timezone (no label). */
    deadlineFormatted() {
      return formatInTimezone(this.responseDeadline, this.eventTimezone)
    },
    /** Live relative countdown to the deadline, e.g. "in 23 minutes" */
    deadlineRelative() {
      return formatTimeUntil(this.responseDeadline, this.nowTs)
    },
    /** Whether availability editing controls should be shown */
    canRespond() {
      return !this.deadlinePassed
    },
    /** Formatted date + time of the persisted scheduled event in the event
     *  timezone, with an explicit TZ label, or "" */
    scheduledEventStartFormatted() {
      const start = this.event?.scheduledEvent?.startDate
      if (!start) return ""
      // Label shows only when the rendered zone differs from the viewer's browser zone.
      return formatInTimezone(start, this.eventTimezone)
    },
    /** Live relative time to the scheduled start, e.g. "in 23 minutes" / "happening now". */
    scheduledRelative() {
      const start = this.event?.scheduledEvent?.startDate
      if (!start) return ""
      const startMs = new Date(start).getTime()
      if (isNaN(startMs)) return ""
      const end = this.event?.scheduledEvent?.endDate
      const endMs = end ? new Date(end).getTime() : startMs + 60 * 60 * 1000
      if (this.nowTs >= startMs && this.nowTs < endMs) return "happening now"
      if (this.nowTs >= endMs) return "ended"
      return formatTimeUntil(start, this.nowTs)
    },
    /** The meeting link, only when it's a non-empty http(s) URL */
    meetingLinkUrl() {
      const link = this.event?.meetingLink
      if (typeof link !== "string" || link.length === 0) return ""
      return /^https?:\/\//i.test(link) ? link : ""
    },
  },

  methods: {
    ...mapActions(["showError", "showInfo", "getEvents"]),
    ...mapMutations(["setAuthUser"]),

    /** Show choice dialog if not signed in, otherwise, immediately start editing availability */
    addAvailability() {
      if (!this.scheduleOverlapComponent) return

      // Don't allow adding/editing availability once the response deadline has passed
      if (this.deadlinePassed && !this.isGroup) {
        this.showInfo("Responses are closed for this event.")
        return
      }

      // Start editing immediately if days only
      if (this.event?.daysOnly) {
        this.scheduleOverlapComponent.startEditing()
        return
      }

      // Start editing if calendar permission granted or user has responded, otherwise show choice dialog
      if (
        (this.authUser && this.calendarPermissionGranted) ||
        this.userHasResponded
      ) {
        this.scheduleOverlapComponent.startEditing()
        if (!this.userHasResponded && !this.isSignUp) {
          this.scheduleOverlapComponent.setAvailabilityAutomatically()
        }
      } else {
        this.choiceDialog = true
      }
    },
    /** Add guest availability while signed in */
    addAvailabilityAsGuest() {
      this.addingAvailabilityAsGuest = true
      this.setAvailabilityManually()
    },
    cancelEditing() {
      /* Cancels editing and resets availability to previous */
      if (!this.scheduleOverlapComponent) return

      if (!this.isSignUp)
        this.scheduleOverlapComponent.resetCurUserAvailability()
      else this.scheduleOverlapComponent.resetSignUpForm()
      this.scheduleOverlapComponent.stopEditing()
      this.curGuestId = ""
      this.addingAvailabilityAsGuest = false
    },
    copyLink() {
      /* Copies event link to clipboard */
      navigator.clipboard.writeText(
        `${window.location.origin}/e/${this.event.shortId ?? this.event._id}`
      )
      this.showInfo("Link copied to clipboard!")
    },
    async deleteAvailability() {
      if (!this.scheduleOverlapComponent) return

      if (!this.authUser || this.addingAvailabilityAsGuest) {
        if (this.curGuestId) {
          await this.scheduleOverlapComponent.deleteAvailability(
            this.curGuestId
          )
          this.curGuestId = ""
        }
      } else {
        await this.scheduleOverlapComponent.deleteAvailability()
      }

      this.showInfo(this.isGroup ? "Left group!" : "Availability deleted!")
      this.scheduleOverlapComponent.stopEditing()
    },

    editEvent() {
      /* Show edit event dialog */
      this.editEventDialog = true
    },
    /** Refresh event details */
    async refreshEvent() {
      let sanitizedId = this.eventId.replaceAll(".", "")

      let resolvedLongId = this.event?._id || ""
      try {
        const ids = await get(`/events/${sanitizedId}/ids`)
        if (ids?.longId) {
          resolvedLongId = ids.longId
        }
      } catch (err) {
        // If ID resolution fails, continue with existing fallback behavior.
      }
      // Try to get guest name from localStorage using resolved longId.
      let guestName = null
      if (typeof localStorage !== "undefined") {
        if (resolvedLongId) {
          guestName = localStorage[`${resolvedLongId}.guestName`]
        }
      }

      // Build URL with guestName if available
      let url = `/events/${sanitizedId}`
      if (guestName && guestName.length > 0) {
        url += `?guestName=${encodeURIComponent(guestName)}`
      }

      // Make single request with guestName if available
      this.event = await get(url)
      processEvent(this.event)
    },

    async checkOwnerPremium() {
      const ownerId = this.event?.ownerId
      if (ownerId && ownerId !== guestUserId) {
        try {
          const res = await get(`/users/${ownerId}/is-premium`)
          this.ownerIsPremium = res.isPremium
        } catch {
          this.ownerIsPremium = false
        }
      }
      this.ownerPremiumChecked = true
    },

    setAvailabilityAutomatically(calendarType = calendarTypes.GOOGLE) {
      /* Prompts user to sign in when "set availability automatically" button clicked */
      if (isWebview(navigator.userAgent)) {
        // Show dialog prompting user to use a real browser
        this.webviewDialog = true
      } else {
        // Or sign in if user is already using a real browser
        let signInParams
        if (this.authUser) {
          // Request permission if calendar permissions not yet granted
          signInParams = {
            state: {
              type: this.isGroup
                ? authTypes.GROUP_ADD_AVAILABILITY
                : authTypes.EVENT_ADD_AVAILABILITY,
              eventId: this.eventId,
            },
            selectAccount: false,
            requestCalendarPermission: true,
          }
        } else {
          // Ask the user to select the account they want to sign in with if not logged in yet
          signInParams = {
            state: {
              type: authTypes.EVENT_ADD_AVAILABILITY,
              eventId: this.eventId,
            },
            selectAccount: true,
            requestCalendarPermission: true,
          }
        }

        if (calendarType === calendarTypes.GOOGLE) {
          signInGoogle(signInParams)
        } else if (calendarType === calendarTypes.OUTLOOK) {
          signInOutlook(signInParams)
        }
      }
      this.choiceDialog = false
    },
    setAvailabilityManually() {
      /* Starts editing after "set availability manually" button clicked */
      if (!this.scheduleOverlapComponent) return

      this.$nextTick(() => {
        this.scheduleOverlapComponent.startEditing()
      })
      this.choiceDialog = false
    },
    editGuestAvailability() {
      /* Edits the selected guest's availability */
      if (!this.scheduleOverlapComponent) return

      this.curGuestId = this.selectedGuestRespondent
      this.scheduleOverlapComponent.startEditing()
      this.$nextTick(() => {
        this.scheduleOverlapComponent.populateUserAvailability(
          this.selectedGuestRespondent
        )
      })
    },

    async saveChanges(ignorePagesNotVisited = false) {
      /* Shows guest dialog if not signed in, otherwise saves auth user's availability */
      if (!this.scheduleOverlapComponent) return

      // If user hasn't responded and they haven't gone to the next page, show pages not visited dialog
      if (
        !this.userHasResponded &&
        this.curGuestId.length === 0 &&
        !this.scheduleOverlapComponent.pageHasChanged &&
        !ignorePagesNotVisited &&
        this.scheduleOverlapComponent.hasPages
      ) {
        this.pagesNotVisitedDialog = true
        return
      }

      if (!this.authUser || this.addingAvailabilityAsGuest) {
        if (this.curGuestId) {
          // Returning guest editing an existing response: they already consented
          // to the privacy policy on their first submission, so re-affirm it.
          this.saveChangesAsGuest({
            name: this.curGuestId,
            email: this.event.responses[this.curGuestId].email,
            consentedToPrivacyPolicy: true,
          })
          this.curGuestId = ""
          this.addingAvailabilityAsGuest = false
        } else {
          this.guestDialog = true
        }
        return
      }

      let changesPersisted = true

      try {
        if (this.isSignUp) {
          changesPersisted =
            await this.scheduleOverlapComponent.submitNewSignUpBlocks()
        } else {
          await this.scheduleOverlapComponent.submitAvailability()
        }
      } catch (err) {
        if (this.isResponseDeadlinePassedError(err)) {
          this.handleResponseDeadlinePassed()
          return
        }
        throw err
      }

      if (changesPersisted) {
        this.showInfo("Changes saved! 💡 Suggest a topic below.")
        this.scheduleOverlapComponent.stopEditing()
        this.promptTopics()
      }
    },
    /** After saving availability, nudge the user toward the Suggest-a-topic box. */
    promptTopics() {
      this.$nextTick(() => {
        setTimeout(() => {
          if (this.$refs.eventTopics && this.$refs.eventTopics.focusSuggest) {
            this.$refs.eventTopics.focusSuggest()
          }
        }, 500)
      })
    },
    async saveChangesAsGuest(payload) {
      /* After guest dialog is submitted, submit availability with the given name */
      if (!this.scheduleOverlapComponent) return

      if (payload.name.length > 0) {
        try {
          await this.scheduleOverlapComponent.submitAvailability(payload)
        } catch (err) {
          if (this.isResponseDeadlinePassedError(err)) {
            this.handleResponseDeadlinePassed()
            return
          }
          throw err
        }

        this.showInfo("Changes saved! 💡 Suggest a topic below.")
        this.scheduleOverlapComponent.resetCurUserAvailability()
        this.scheduleOverlapComponent.stopEditing()
        this.promptTopics()
        this.guestDialog = false
        this.addingAvailabilityAsGuest = false
      }
    },
    /** Detects the backend's 403 response-deadline-passed error */
    isResponseDeadlinePassedError(err) {
      return (
        err?.status === 403 &&
        err?.parsed?.error === "response-deadline-passed"
      )
    },
    /** Shows a friendly message and exits editing when responses are closed */
    handleResponseDeadlinePassed() {
      this.showInfo("Responses are closed for this event.")
      // Refresh so the deadline banner / closed state reflects reality
      this.refreshEvent()
      this.scheduleOverlapComponent?.stopEditing()
      this.guestDialog = false
      this.addingAvailabilityAsGuest = false
    },

    /** Owner saved a final schedule via the dialog: update locally + re-fetch */
    /** Open the schedule dialog normally (no prefill from the grid). */
    openScheduleDialog() {
      this.schedulePrefill = null
      this.scheduleEventDialog = true
    },
    /** Open the schedule dialog prefilled with a date/time chosen via the grid
     *  "schedule event" flow, so the owner only needs to add the Meet link. */
    openScheduleDialogPrefilled(payload) {
      this.schedulePrefill = payload
      this.scheduleEventDialog = true
    },
    onEventScheduled(scheduledEvent) {
      if (this.event) {
        this.$set(this.event, "scheduledEvent", {
          summary: scheduledEvent.summary,
          startDate: scheduledEvent.startDate,
          endDate: scheduledEvent.endDate,
        })
        this.$set(this.event, "meetingLink", scheduledEvent.meetingLink || "")
      }
      this.showInfo("Event scheduled!")
      this.refreshEvent()
    },
    /** Owner cleared the schedule via the dialog: update locally + re-fetch */
    onEventScheduleCleared() {
      if (this.event) {
        this.$set(this.event, "scheduledEvent", null)
        this.$set(this.event, "meetingLink", "")
      }
      this.showInfo("Schedule cleared.")
      this.refreshEvent()
    },

    scheduleEvent() {
      this.scheduleOverlapComponent?.scheduleEvent()
    },
    cancelScheduleEvent() {
      this.scheduleOverlapComponent?.cancelScheduleEvent()
    },
    confirmScheduleEvent() {
      this.scheduleOverlapComponent?.confirmScheduleEvent()
    },

    highlightAvailabilityBtn() {
      // if (!this.isPhone) {
      //   window.scrollTo({ top: 0, behavior: "instant" })
      // }
      this.availabilityBtnOpacity = 0.1
      setTimeout(() => {
        this.availabilityBtnOpacity = 1
        setTimeout(() => {
          this.availabilityBtnOpacity = 0.1
          setTimeout(() => {
            this.availabilityBtnOpacity = 1
          }, 100)
        }, 100)
      }, 100)
    },

    /** Sign in with google to link apple calendar */
    signInLinkApple() {
      if (isWebview(navigator.userAgent)) {
        // Show dialog prompting user to use a real browser
        this.webviewDialog = true
      } else {
        signInGoogle({
          state: {
            type: authTypes.EVENT_SIGN_IN_LINK_APPLE,
            eventId: this.eventId,
          },
          selectAccount: true,
        })
      }
    },
    /** Called when user adds apple calendar account */
    addedAppleCalendar() {
      this.choiceDialog = false
      this.scheduleOverlapComponent?.startEditing()
      this.scheduleOverlapComponent?.setAvailabilityAutomatically()
    },
    /** Called when user adds ICS calendar account */
    addedICSCalendar() {
      this.choiceDialog = false
      this.scheduleOverlapComponent?.startEditing()
      this.scheduleOverlapComponent?.setAvailabilityAutomatically()
    },

    /** Refresh calendar availabilities of everybody in the group */
    async fetchCalendarAvailabilities() {
      if (this.event.type !== eventTypes.GROUP) return

      // this.calendarAvailabilities = {}
      const curWeekOffset = this.weekOffset
      return getCalendarEventsMap(this.event, {
        weekOffset: curWeekOffset,
        eventId: this.event._id,
      })
        .then((calendarAvailabilities) => {
          // Don't update calendar availabilities if user
          // selected a different weekoffset by the time these calendar events load
          if (curWeekOffset !== this.weekOffset) return

          this.calendarAvailabilities = calendarAvailabilities

          // Fix DST bug
          for (const userId in this.calendarAvailabilities) {
            for (const index in this.calendarAvailabilities[userId]) {
              const event = this.calendarAvailabilities[userId][index]
              const startDate = new Date(event.startDate)
              const endDate = new Date(event.endDate)
              if (doesDstExist(startDate) && !isDstObserved(startDate)) {
                startDate.setHours(startDate.getHours() - 1)
                endDate.setHours(endDate.getHours() - 1)
              }
              this.calendarAvailabilities[userId][index].startDate =
                startDate.toISOString()
              this.calendarAvailabilities[userId][index].endDate =
                endDate.toISOString()
            }
          }
        })
        .catch((err) => {
          console.error(err)
        })
    },

    /** Fetch current user's calendar events */
    async fetchAuthUserCalendarEvents() {
      if (!this.authUser) {
        this.calendarPermissionGranted = false
        return
      }

      // this.calendarEventsMap = {}
      const curWeekOffset = this.weekOffset
      return getCalendarEventsMap(this.event, { weekOffset: curWeekOffset })
        .then((eventsMap) => {
          // If all calendars have error, then set calendarPermissionGranted to false
          // TODO: What happens if user signed in without enabling calendar??
          // let noError = false
          // for (const key in eventsMap) {
          //   if (!eventsMap[key].error) {
          //     noError = true
          //     break
          //   }
          // }
          // if (!noError) {
          //   this.calendarPermissionGranted = false
          //   return
          // }

          // Don't set calendar events / set availability if user has already
          // selected a different weekoffset by the time these calendar events load
          if (curWeekOffset !== this.weekOffset) return

          this.calendarEventsMap = eventsMap

          // Fix DST bug
          if (
            this.event.type === eventTypes.GROUP ||
            this.event.type === eventTypes.DOW
          ) {
            for (const calendarId in this.calendarEventsMap) {
              for (const index in this.calendarEventsMap[calendarId]
                .calendarEvents) {
                const event =
                  this.calendarEventsMap[calendarId].calendarEvents[index]
                const startDate = new Date(event.startDate)
                const endDate = new Date(event.endDate)
                if (doesDstExist(startDate) && !isDstObserved(startDate)) {
                  startDate.setHours(startDate.getHours() - 1)
                  endDate.setHours(endDate.getHours() - 1)
                }
                this.calendarEventsMap[calendarId].calendarEvents[
                  index
                ].startDate = startDate.toISOString()
                this.calendarEventsMap[calendarId].calendarEvents[
                  index
                ].endDate = endDate.toISOString()
              }
            }
          }

          // Set user availability automatically if we're in editing mode and they haven't responded
          if (
            this.authUser &&
            this.isEditing &&
            !this.userHasResponded &&
            !this.areUnsavedChanges &&
            this.scheduleOverlapComponent
          ) {
            this.$nextTick(() => {
              this.scheduleOverlapComponent?.setAvailabilityAutomatically()
            })
          }

          // calendar permission granted is false when every calendar in the calendar map has an error, true otherwise
          this.calendarPermissionGranted = !Object.values(
            this.calendarEventsMap
          ).every((c) => Boolean(c.error))

          if (!this.hasRefetchedAuthUserCalendarEvents) {
            const hasAtLeastOneError = Object.values(
              this.calendarEventsMap
            ).some((c) => Boolean(c.error))

            // Refetch calendar if there is an error
            if (hasAtLeastOneError) {
              this.hasRefetchedAuthUserCalendarEvents = true
              setTimeout(() => {
                this.fetchAuthUserCalendarEvents()
              }, 1000)
            }
          }
        })
        .catch((err) => {
          console.error(err)
          this.calendarPermissionGranted = false
        })
    },

    /** Refreshes calendar avaliabilities and fetches current user calendar events */
    refreshCalendar() {
      const promises = []
      promises.push(this.fetchCalendarAvailabilities())
      promises.push(this.fetchAuthUserCalendarEvents())

      const curWeekOffset = this.weekOffset
      this.loading = true
      Promise.allSettled(promises).then(() => {
        // Only set loading to false if promises resolved at the same week offset they were fetched at
        // i.e. no new promises are currently being run
        if (curWeekOffset === this.weekOffset) {
          this.loading = false
        }
      })
    },

    /** Resets week offset to 0 */
    resetWeekOffset() {
      if (this.event && this.event.startOnMonday) {
        this.weekOffset = 0
      } else {
        this.weekOffset = 0
      }
    },

    onBeforeUnload(e) {
      if (this.areUnsavedChanges) {
        e.preventDefault()
        e.returnValue = ""
        return
      }

      delete e["returnValue"]
    },

    handleGuestDialogSubmit(guestPayload) {
      this.saveChangesAsGuest(guestPayload)
    },


    // -----------------------------------
    //#region Sign Up Form
    // -----------------------------------

    initiateSignUpFlow(signUpBlock) {
      this.currSignUpBlock = signUpBlock
      this.signUpForSlotDialog = true
    },

    async signUpForBlock(guestPayload) {
      let payload

      if (this.authUser) {
        payload = {
          guest: false,
          signUpBlockIds: [this.currSignUpBlock._id],
        }
      } else {
        payload = {
          guest: true,
          signUpBlockIds: [this.currSignUpBlock._id],
          ...guestPayload,
        }
      }

      await post(`/events/${this.event._id}/response`, payload)
      await this.refreshEvent()

      this.scheduleOverlapComponent.resetSignUpForm()
      this.signUpForSlotDialog = false
    },

    //#endregion
  },

  async created() {
    window.addEventListener("beforeunload", this.onBeforeUnload)

    // Get event details
    try {
      await this.refreshEvent()
      await this.checkOwnerPremium()

      // Redirect if we're at the wrong route
      if (this.event.type === eventTypes.GROUP) {
        if (this.$route.name === "event") {
          this.$router.replace({
            name: "group",
            params: {
              groupId: this.eventId,
              initialTimezone: this.initialTimezone,
              fromSignIn: this.fromSignIn,
              contactsPayload: this.contactsPayload,
            },
          })
        }
      } else {
        if (this.$route.name === "group") {
          this.$router.replace({
            name: "event",
            params: {
              eventId: this.eventId,
              initialTimezone: this.initialTimezone,
              fromSignIn: this.fromSignIn,
              contactsPayload: this.contactsPayload,
            },
          })
        }
      }

      const fromEditEvent = localStorage.getItem(
        `from-edit-event-${this.event._id}`
      )
      if (fromEditEvent) {
        localStorage.removeItem(`from-edit-event-${this.event._id}`)
        this.fromEditEvent = true
      }
    } catch (err) {
      switch (err.error) {
        case errors.EventNotFound:
          this.showError("The specified event does not exist!")
          this.$router.replace({ name: "home" })
          return
      }
    }

    const promises = []
    promises.push(this.fetchCalendarAvailabilities())
    promises.push(this.fetchAuthUserCalendarEvents())

    this.loading = true
    Promise.allSettled(promises).then(() => {
      this.loading = false
    })

    get("/user/profile")
      .then((authUser) => {
        this.setAuthUser(authUser)
      })
      .catch(() => {
        this.setAuthUser(null)
      })
  },

  beforeDestroy() {
    if (this.deadlineTimer) clearInterval(this.deadlineTimer)
    window.removeEventListener("beforeunload", this.onBeforeUnload)
  },

  watch: {
    event() {
      if (this.event) {
        this.resetWeekOffset()
        this.$nextTick(() => {
          this.scheduleOverlapComponent = this.$refs.scheduleOverlap
        })
        document.title = `${this.event.name} - WannPassts`
      }
    },
    scheduleOverlapComponent() {
      if (!this.scheduleOverlapComponentLoaded) {
        this.scheduleOverlapComponentLoaded = true

        // Put into editing mode if just signed in
        if ((this.fromSignIn || this.editingMode) && !this.isGroup) {
          this.scheduleOverlapComponent.startEditing()
        }

        if (this.isGroup && !this.userHasResponded) {
          this.invitationDialog = true
        }
      }
    },
    weekOffset() {
      this.refreshCalendar()
    },
    [`authUser.calendarAccounts`]() {
      this.fetchAuthUserCalendarEvents()
    },
  },
}
</script>
