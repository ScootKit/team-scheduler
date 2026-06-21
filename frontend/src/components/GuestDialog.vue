<template>
  <v-dialog
    :value="value"
    @input="(e) => $emit('input', e)"
    width="400"
    content-class="tw-m-0"
  >
    <v-card>
      <v-card-title class="tw-flex">
        <div>Continue as guest</div>
        <v-spacer />
        <v-btn icon @click="$emit('input', false)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-card-title>
      <v-card-text>
        <v-form
          ref="form"
          v-model="formValid"
          lazy-validation
          class="tw-flex tw-flex-col tw-gap-y-4"
          onsubmit="return false;"
        >
          <v-text-field
            v-model="name"
            @keyup.enter="submit"
            :rules="nameRules"
            placeholder="Enter your name..."
            :autofocus="$autofocusEnabled"
            hide-details="auto"
            autocomplete="off"
            solo
          ></v-text-field>
          <v-text-field
            v-if="event.collectEmails"
            v-model="email"
            @keyup.enter="submit"
            :rules="emailRules"
            placeholder="Enter your email..."
            hint="The event creator has requested your email. It will only be visible to them."
            persistent-hint
            solo
          ></v-text-field>
          <v-checkbox
            v-model="consent"
            :rules="consentRules"
            hide-details="auto"
            class="tw-mt-0 tw-pt-0"
          >
            <template #label>
              <span class="tw-text-sm">
                I agree to the
                <a
                  v-if="privacyPolicyUrl"
                  :href="privacyPolicyUrl"
                  target="_blank"
                  rel="noopener"
                  @click.stop
                  >Privacy Policy</a
                >
                <template v-else>Privacy Policy</template>
              </span>
            </template>
          </v-checkbox>
          <div class="tw-flex">
            <v-spacer />
            <v-btn
              @click="submit"
              class="tw-bg-green"
              :dark="formValid && consent"
              :disabled="!formValid || !consent"
            >
              Continue
            </v-btn>
          </div>
          <div class="tw-mt-1 tw-text-center tw-text-xs tw-text-dark-gray">
            ScootKit employee or helper?
            <router-link :to="{ name: 'sign-in' }" class="tw-text-blue">
              Sign in to your account
            </router-link>
            instead.
          </div>
        </v-form>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { isPhone, validateEmail } from "@/utils"
import { privacyPolicyUrl } from "@/constants"

export default {
  name: "GuestDialog",

  emits: ["input", "submit"],

  props: {
    value: { type: Boolean, required: true },
    event: { type: Object, required: true },
    respondents: { type: Array, required: true },
  },

  data() {
    return {
      formValid: false,
      name: "",
      email: "",
      consent: false,
      nameRules: [],
      emailRules: [],
      consentRules: [],
    }
  },

  computed: {
    isPhone() {
      return isPhone(this.$vuetify)
    },
    privacyPolicyUrl: () => privacyPolicyUrl,
  },

  methods: {
    submit() {
      // Set rules only on submit
      this.nameRules = [
        (name) => !!name || "Name is required",
        (name) => !this.respondents.includes(name) || "Name already taken",
      ]
      this.emailRules = [
        (email) => !!email || "Email is required",
        (email) => !!validateEmail(email) || "Invalid email",
      ]
      this.consentRules = [
        (consent) => !!consent || "You must agree to the Privacy Policy",
      ]

      this.$nextTick(() => {
        if (!this.$refs.form.validate()) return

        this.$emit("submit", {
          name: this.name,
          email: this.email,
          consentedToPrivacyPolicy: this.consent,
        })
      })
    },
  },

  watch: {
    value() {
      if (this.value) {
        this.name = ""
        this.email = ""
        this.consent = false
        this.nameRules = []
        this.emailRules = []
        this.consentRules = []

        this.$refs.form?.resetValidation()
      }
    },
    name() {
      // Default rules before submitting
      this.nameRules = [
        (name) => !this.respondents.includes(name) || "Name already taken",
      ]
    },
    email() {
      // Default rules before submitting
      this.emailRules = []
    },
  },
}
</script>
