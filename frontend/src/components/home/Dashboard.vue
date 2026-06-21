<template>
  <div class="tw-rounded-md tw-px-6 tw-py-4 sm:tw-mx-4 sm:tw-bg-[#f3f3f366]">
    <div class="tw-mb-3 tw-flex tw-items-center tw-justify-between">
      <div class="tw-flex tw-flex-col">
        <div
          class="tw-text-xl tw-font-medium tw-text-dark-green sm:tw-text-2xl"
        >
          Dashboard
        </div>
        <div
          v-if="!isPremiumUser"
          class="tw-flex tw-items-baseline tw-gap-2 tw-text-sm tw-font-normal tw-text-very-dark-gray"
        >
          <div>
            {{ authUser?.numEventsCreated }} / {{ numFreeEvents }} free events
            created this month
          </div>
          <div
            class="tw-cursor-pointer tw-select-none tw-text-xs tw-font-medium tw-text-green tw-underline"
            @click="openUpgradeDialog"
          >
            Upgrade
          </div>
        </div>
      </div>
      <div class="tw-flex tw-items-center tw-gap-1">
        <v-btn
          text
          small
          @click="hidePastEvents = !hidePastEvents"
          class="tw-text-very-dark-gray"
        >
          <v-icon class="tw-text-lg">{{
            hidePastEvents ? "mdi-eye-off-outline" : "mdi-eye-outline"
          }}</v-icon>
          <span class="tw-ml-2">{{
            hidePastEvents ? "Show past" : "Hide past"
          }}</span>
        </v-btn>
        <v-btn
          text
          @click="openCreateFolderDialog"
          class="tw-text-very-dark-gray"
        >
          <v-icon class="tw-text-lg">mdi-folder-plus</v-icon>
          <span class="tw-ml-2">New folder</span>
        </v-btn>
      </div>
    </div>

    <div>
      <div
        v-for="folder in allFolders"
        :key="folder.id"
        class="tw-group tw-mb-2"
      >
        <div class="tw-flex tw-items-center">
          <v-btn icon small @click="toggleFolder(folder.id)">
            <v-icon>{{
              folderOpenState[folder.id] ? "mdi-menu-down" : "mdi-menu-right"
            }}</v-icon>
          </v-btn>
          <v-chip
            v-if="folder.type === 'regular'"
            :color="folder.color || '#D3D3D3'"
            small
            class="tw-mr-2 tw-rounded tw-border tw-border-light-gray-stroke tw-px-2 tw-text-sm tw-font-medium"
            :class="folder.isOwn ? 'tw-cursor-pointer' : ''"
            @click="folder.isOwn && openEditFolderDialog(folder)"
          >
            {{ folder.name }}
          </v-chip>
          <span v-else class="tw-mr-2 tw-text-sm tw-font-medium">{{
            folder.name
          }}</span>
          <v-chip
            v-if="folder.type === 'regular' && folder.isPublic"
            x-small
            class="tw-mr-2 tw-bg-light-gray tw-text-dark-gray"
            title="This folder is public and visible to everyone"
          >
            <v-icon x-small left>mdi-earth</v-icon>
            Public
          </v-chip>
          <div
            v-if="folder.type === 'regular' && folder.isOwn"
            class="tw-invisible tw-flex tw-items-center group-hover:tw-visible"
          >
            <v-menu offset-y>
              <template v-slot:activator="{ on, attrs }">
                <v-btn icon small v-bind="attrs" v-on="on" @click.stop.prevent>
                  <v-icon small>mdi-dots-horizontal</v-icon>
                </v-btn>
              </template>
              <v-list dense class="tw-py-1">
                <v-list-item @click.stop.prevent="openEditFolderDialog(folder)">
                  <v-list-item-title>Edit</v-list-item-title>
                </v-list-item>
                <v-list-item @click.stop.prevent="openDeleteDialog(folder)">
                  <v-list-item-title class="tw-text-red"
                    >Delete</v-list-item-title
                  >
                </v-list-item>
              </v-list>
            </v-menu>
            <v-btn
              icon
              small
              @click.stop.prevent="createEventInFolder(folder.id)"
            >
              <v-icon small>mdi-plus</v-icon>
            </v-btn>
          </div>
        </div>
        <div v-show="folderOpenState[folder.id]">
          <draggable
            :list="[
              ...eventsByFolder[folder.id].groups,
              ...eventsByFolder[folder.id].events,
            ]"
            :group="
              folder.type === 'regular' && !folder.isOwn
                ? { name: 'events', pull: false, put: false }
                : 'events'
            "
            :disabled="folder.type === 'regular' && !folder.isOwn"
            @end="onEnd"
            :data-folder-id="
              folder.type === 'no-folder'
                ? 'null'
                : folder.type === 'archived'
                ? 'archived'
                : folder.id
            "
            draggable=".item"
            :delay="200"
            :delay-on-touch-only="true"
            :class="[
              'tw-relative tw-grid tw-min-h-[52px] tw-grid-cols-1 tw-gap-4 tw-py-4 sm:tw-grid-cols-2',
              folder.type === 'archived' ? 'tw-opacity-75' : '',
            ]"
          >
            <template v-slot:header>
              <div
                v-if="
                  eventsByFolder[folder.id].groups.length === 0 &&
                  eventsByFolder[folder.id].events.length === 0
                "
                class="tw-absolute tw-left-0 tw-py-4 tw-text-sm tw-text-very-dark-gray"
                :class="folder.type === 'regular' ? 'tw-ml-8' : 'tw-ml-7'"
              >
                {{ folder.emptyMessage }}
              </div>
            </template>
            <template v-if="eventsByFolder[folder.id].groups.length > 0">
              <EventItem
                v-for="event in eventsByFolder[folder.id].groups"
                :key="event._id"
                :id="event._id"
                :event="event"
                :folder-id="folder.id"
                class="item"
              />
              <div class="tw-col-span-full"></div>
            </template>
            <EventItem
              v-for="event in eventsByFolder[folder.id].events"
              :key="event._id"
              :id="event._id"
              :event="event"
              :folder-id="folder.id"
              class="item"
            />
          </draggable>
        </div>
      </div>

      <div v-if="allEvents.length === 0">
        <div class="tw-py-4 tw-text-sm tw-text-very-dark-gray">
          No events yet! Create one to get started.
        </div>
      </div>
    </div>
    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ folderToDelete.name }}"?</v-card-title>
        <v-card-text
          >Are you sure you want to delete this folder? All events you own in
          this folder will be deleted as well.</v-card-text
        >
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="red darken-1" text @click="confirmDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-dialog v-model="createFolderDialog" max-width="400">
      <v-card>
        <v-card-title>{{ folderDialogTitle }}</v-card-title>
        <v-card-text>
          <v-text-field
            v-model="newFolderName"
            label="Folder name"
            placeholder="Untitled folder"
            :autofocus="$autofocusEnabled"
            @keydown.enter="confirmFolderDialog"
            hide-details
          ></v-text-field>
          <div class="tw-mt-4">
            <span class="tw-text-gray-500 tw-text-sm">Color</span>
            <div class="tw-mt-2 tw-flex tw-gap-x-3">
              <div
                v-for="color in folderColors"
                :key="color"
                class="tw-h-6 tw-w-6 tw-cursor-pointer tw-rounded-full tw-border tw-border-light-gray-stroke"
                :style="{ backgroundColor: color }"
                :class="{
                  'tw-ring-2 tw-ring-gray tw-ring-offset-2':
                    newFolderColor === color,
                }"
                @click="newFolderColor = color"
              ></div>
            </div>
          </div>
          <div class="tw-mt-5">
            <v-text-field
              v-model="newFolderWebhookUrl"
              label="Discord webhook URL (optional)"
              placeholder="https://discord.com/api/webhooks/..."
              hide-details="auto"
              clearable
            ></v-text-field>
            <div class="tw-mt-1 tw-text-xs tw-text-dark-gray">
              Posts to this Discord channel when an event is added to this folder
              or scheduled.
            </div>
          </div>
          <div v-if="authUser && authUser.isAdmin" class="tw-mt-4">
            <v-switch
              v-model="newFolderIsPublic"
              label="Make this folder public (visible to everyone)"
              hide-details
              class="tw-mt-0"
            ></v-switch>
          </div>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="closeFolderDialog">Cancel</v-btn>
          <v-btn color="primary" text @click="confirmFolderDialog">{{
            folderDialogConfirmText
          }}</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { mapState, mapActions, mapGetters } from "vuex"
import draggable from "vuedraggable"
import {
  eventTypes,
  folderColors,
  numFreeEvents,
  upgradeDialogTypes,
} from "@/constants"
import EventItem from "@/components/EventItem.vue"
import ObjectID from "bson-objectid"

export default {
  name: "Dashboard",
  components: {
    EventItem,
    draggable,
  },
  data() {
    return {
      deleteDialog: false,
      folderToDelete: {},
      hidePastEvents: localStorage["hidePastEvents"] === "true",
      createFolderDialog: false,
      newFolderName: "",
      newFolderColor: folderColors[3],
      newFolderWebhookUrl: "",
      newFolderIsPublic: false,
      isEditingFolder: false,
      folderToEdit: null,
      folderOpenState: {
        "no-folder": true,
      },
    }
  },
  computed: {
    ...mapGetters(["isPremiumUser"]),
    ...mapState(["authUser", "events", "groupsEnabled", "folders"]),
    orderedFolders() {
      return this.folders.sort((a, b) => {
        return a.name.localeCompare(b.name)
      })
    },
    numFreeEvents() {
      return numFreeEvents
    },
    folderColors() {
      return folderColors
    },
    allEvents() {
      return this.events
    },
    allEventsMap() {
      return this.allEvents.reduce((acc, event) => {
        acc[event._id] = event
        return acc
      }, {})
    },
    eventsByFolder() {
      const eventsByFolder = {}
      const allEventIds = new Set(this.allEvents.map((e) => e._id))

      eventsByFolder["no-folder"] = { groups: [], events: [] }
      eventsByFolder["archived"] = { groups: [], events: [] }

      this.folders.forEach((folder) => {
        eventsByFolder[folder._id] = { groups: [], events: [] }
        for (const eventId of folder.eventIds) {
          const event = this.allEventsMap[eventId]
          if (event) {
            // When hiding past events, skip them (but still claim the id so it
            // doesn't fall through to the no-folder bucket).
            if (this.hidePastEvents && this.isEventPast(event)) {
              allEventIds.delete(eventId)
              continue
            }
            if (event.isArchived) {
              if (event.type === eventTypes.GROUP) {
                eventsByFolder["archived"].groups.push(event)
              } else {
                eventsByFolder["archived"].events.push(event)
              }
            } else {
              if (event.type === eventTypes.GROUP) {
                eventsByFolder[folder._id].groups.push(event)
              } else {
                eventsByFolder[folder._id].events.push(event)
              }
            }
            allEventIds.delete(eventId)
          }
        }
        eventsByFolder[folder._id].groups.sort(this.sortEvents)
        eventsByFolder[folder._id].events.sort(this.sortEvents)
      })

      for (const eventId of allEventIds) {
        const event = this.allEventsMap[eventId]
        if (event) {
          if (this.hidePastEvents && this.isEventPast(event)) {
            continue
          }
          if (event.isArchived) {
            if (event.type === eventTypes.GROUP) {
              eventsByFolder["archived"].groups.push(event)
            } else {
              eventsByFolder["archived"].events.push(event)
            }
          } else {
            if (event.type === eventTypes.GROUP) {
              eventsByFolder["no-folder"].groups.push(event)
            } else {
              eventsByFolder["no-folder"].events.push(event)
            }
          }
        }
      }

      eventsByFolder["no-folder"].groups.sort(this.sortEvents)
      eventsByFolder["no-folder"].events.sort(this.sortEvents)
      eventsByFolder["archived"].groups.sort(this.sortEvents)
      eventsByFolder["archived"].events.sort(this.sortEvents)
      return eventsByFolder
    },
    folderDialogTitle() {
      return this.isEditingFolder ? "Edit folder" : "New folder"
    },
    folderDialogConfirmText() {
      return this.isEditingFolder ? "Save" : "Create"
    },
    allFolders() {
      const folders = this.folders.map((folder) => ({
        ...folder,
        id: folder._id,
        type: "regular",
        name: folder.name,
        userId: folder.userId,
        isPublic: folder.isPublic,
        isOwn: this.isOwnFolder(folder),
        emptyMessage: "No events in this folder",
      }))

      // Only show "no-folder" section if there are events
      if (this.allEvents.length > 0) {
        folders.push({
          id: "no-folder",
          type: "no-folder",
          name: "No folder",
          emptyMessage: "No events",
        })
      }

      // Only show "archived" section if there are archived events
      if (this.allEvents.some((event) => event.isArchived)) {
        folders.push({
          id: "archived",
          type: "archived",
          name: "Archived",
          emptyMessage: "No archived events",
        })
      }

      return folders
    },
  },

  methods: {
    ...mapActions([
      "createFolder",
      "showUpgradeDialog",
      "deleteFolder",
      "setEventFolder",
      "updateFolder",
      "createNew",
    ]),
    isOwnFolder(folder) {
      // A regular folder is owned by the current user when its userId matches.
      // Special folders (no-folder/archived) have no userId and are always "own".
      if (!folder || folder.userId === undefined || folder.userId === null) {
        return true
      }
      return folder.userId === this.authUser?._id
    },
    /** Whether an event is in the past (scheduled time elapsed, or all candidate dates passed).
     *  Weekly/day-of-week and group events have no concrete date, so are never past. */
    isEventPast(event) {
      if (!event || event.type === eventTypes.GROUP || event.type === eventTypes.DOW) {
        return false
      }
      const scheduledStart = event.scheduledEvent?.startDate
      if (scheduledStart) {
        const d = new Date(scheduledStart)
        if (!isNaN(d.getTime())) return d.getTime() < Date.now()
      }
      const dates = event.dates
      if (Array.isArray(dates) && dates.length > 0) {
        const latest = Math.max(...dates.map((x) => new Date(x).getTime()))
        if (!isNaN(latest)) {
          // Treat as past once the day after the latest candidate date has begun.
          return latest + 24 * 60 * 60 * 1000 < Date.now()
        }
      }
      return false
    },
    sortEvents(a, b) {
      // Past events sort after upcoming ones.
      const pastA = this.isEventPast(a)
      const pastB = this.isEventPast(b)
      if (pastA !== pastB) return pastA ? 1 : -1
      if (ObjectID.isValid(a._id) && ObjectID.isValid(b._id)) {
        return ObjectID(b._id).getTimestamp() - ObjectID(a._id).getTimestamp()
      }
      return 0
    },
    onEnd(evt) {
      const eventId = evt.item.id
      let newFolderId = evt.to.dataset.folderId
      if (
        newFolderId === "null" ||
        newFolderId === undefined ||
        newFolderId === "no-folder"
      ) {
        newFolderId = null
      }

      // Don't allow dropping into archived section
      if (newFolderId === "archived") {
        return
      }

      let fromFolderId = evt.from.dataset.folderId
      if (fromFolderId === "no-folder") {
        fromFolderId = null
      }
      if (fromFolderId === "archived") {
        fromFolderId = null
      }

      // if moving within the same folder, do nothing.
      if (fromFolderId === newFolderId) {
        // Here you might want to handle re-ordering within the same folder
        // For now, we do nothing.
        return
      }

      const event = this.allEvents.find((e) => e._id === eventId)

      if (event) {
        this.setEventFolder({
          eventId: event._id,
          folderId: newFolderId,
        })
      }
    },
    async confirmFolderDialog() {
      if (!this.newFolderName.trim()) {
        this.closeFolderDialog()
        return
      }
      // Send null to clear the webhook; trimmed string otherwise.
      const webhookUrl = (this.newFolderWebhookUrl || "").trim() || null
      let ok = false
      if (this.isEditingFolder) {
        ok = await this.updateFolder({
          folderId: this.folderToEdit._id,
          name: this.newFolderName.trim(),
          color: this.newFolderColor,
          webhookUrl,
          isPublic: this.newFolderIsPublic,
        })
      } else {
        this.$posthog.capture("folder_created", {
          folderName: this.newFolderName.trim(),
          folderColor: this.newFolderColor,
        })
        ok = await this.createFolder({
          name: this.newFolderName.trim(),
          color: this.newFolderColor,
          webhookUrl,
          isPublic: this.newFolderIsPublic,
        })
      }
      // Keep the dialog open on failure (e.g. invalid Discord webhook URL) so the
      // user can fix the value; the store surfaces the error message.
      if (ok) this.closeFolderDialog()
    },
    closeFolderDialog() {
      this.createFolderDialog = false
      this.isEditingFolder = false
      this.folderToEdit = null
      this.newFolderName = ""
      this.newFolderColor = folderColors[3]
      this.newFolderWebhookUrl = ""
      this.newFolderIsPublic = false
    },
    openCreateFolderDialog() {
      this.isEditingFolder = false
      this.folderToEdit = null
      this.newFolderName = ""
      this.newFolderColor = folderColors[3]
      this.newFolderWebhookUrl = ""
      this.newFolderIsPublic = false
      this.createFolderDialog = true
    },
    openEditFolderDialog(folder) {
      // Don't allow editing folders the current user doesn't own (public folders
      // shared by others are read-only).
      if (!this.isOwnFolder(folder)) {
        return
      }
      this.isEditingFolder = true
      this.folderToEdit = folder
      this.newFolderName = folder.name
      this.newFolderColor = folder.color || folderColors[3]
      this.newFolderWebhookUrl = folder.webhookUrl || ""
      this.newFolderIsPublic = !!folder.isPublic
      this.createFolderDialog = true
    },
    toggleFolder(folderId) {
      this.$set(this.folderOpenState, folderId, !this.folderOpenState[folderId])
    },
    createEventInFolder(folderId) {
      const actualFolderId = folderId === "no-folder" ? null : folderId
      this.createNew({
        eventOnly: false,
        folderId: actualFolderId,
      })
    },
    openDeleteDialog(folder) {
      this.folderToDelete = folder
      this.deleteDialog = true
    },
    confirmDelete() {
      this.$store.dispatch("deleteFolder", this.folderToDelete._id)
      this.deleteDialog = false
    },
    openUpgradeDialog() {
      this.showUpgradeDialog({
        type: upgradeDialogTypes.UPGRADE_MANUALLY,
      })
    },
  },
  created() {
    try {
      const storedState = localStorage.getItem("folderOpenState")
      if (storedState) {
        this.folderOpenState = JSON.parse(storedState)
      }
    } catch (e) {
      console.error("Error reading folderOpenState from localStorage", e)
      // If corrupted, remove it
      localStorage.removeItem("folderOpenState")
    }
  },
  watch: {
    hidePastEvents(val) {
      localStorage["hidePastEvents"] = val ? "true" : "false"
    },
    folderOpenState: {
      handler(newState) {
        try {
          localStorage.setItem("folderOpenState", JSON.stringify(newState))
        } catch (e) {
          console.error("Error saving folderOpenState to localStorage", e)
        }
      },
      deep: true,
    },
    folders: {
      handler(newFolders) {
        if (newFolders) {
          newFolders.forEach((folder) => {
            if (this.folderOpenState[folder._id] === undefined) {
              this.$set(this.folderOpenState, folder._id, true) // default to open
            }
          })
        }
      },
      immediate: true,
    },
  },
}
</script>

<style>
.v-expansion-panel-header {
  padding: 16px 4px !important;
}
</style>
