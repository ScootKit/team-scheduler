import Vue from "vue"
import VueRouter from "vue-router"
import { get } from "@/utils"

Vue.use(VueRouter)

const routes = [
  {
    // Internal tool: there is no public landing page. "/" (and any lingering
    // { name: "landing" } links) redirect to sign-in; the global guard then
    // sends already-authenticated users on to /home.
    path: "/",
    name: "landing",
    redirect: { name: "sign-in" },
  },
  {
    path: "/home",
    name: "home",
    component: () => import("@/views/Home.vue"),
    props: true,
  },
  {
    path: "/settings",
    name: "settings",
    component: () => import("@/views/Settings.vue"),
  },
  {
    path: "/e/:eventId",
    name: "event",
    component: () => import("@/views/Event.vue"),
    props: true,
  },
  {
    path: "/e/:eventId/responded",
    name: "responded",
    component: () => import("@/views/Responded.vue"),
    props: true,
  },
  {
    path: "/g/:groupId",
    name: "group",
    component: () => import("@/views/Group.vue"),
    props: true,
  },
  {
    path: "/s/:signUpId",
    name: "signUp",
    component: () => import("@/views/SignUp.vue"),
    props: true,
  },
  {
    path: "/sign-in",
    name: "sign-in",
    component: () => import("@/views/SignIn.vue"),
  },
  {
    path: "/sign-up",
    name: "sign-up",
    component: () => import("@/views/SignIn.vue"),
    props: { initialIsSignUp: true },
  },
  {
    path: "/auth",
    name: "auth",
    component: () => import("@/views/Auth.vue"),
  },
  {
    path: "/privacy-policy",
    name: "privacy-policy",
    component: () => import("@/views/PrivacyPolicy.vue"),
  },
  {
    path: "/cookie-settings",
    name: "cookie-settings",
    component: () => import("@/components/CookieSettings.vue"),
  },
  {
    path: "/stripe-redirect",
    name: "stripe-redirect",
    component: () => import("@/views/StripeRedirect.vue"),
  },
  {
    path: "/test",
    name: "test",
    component: () => import("@/views/Test.vue"),
  },
  {
    path: "*",
    name: "404",
    component: () => import("@/views/PageNotFound.vue"),
  },
]

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes,
})

router.beforeEach(async (to, from, next) => {
  const authRoutes = ["home", "settings"]
  const noAuthRoutes = ["sign-in", "sign-up"]
  try {
    await get("/auth/status")

    if (noAuthRoutes.includes(to.name)) {
      next({ name: "home" })
    } else {
      next()
    }
  } catch (err) {
    // Not signed in: gated routes redirect to sign-in (there is no landing page).
    if (authRoutes.includes(to.name)) {
      next({ name: "sign-in" })
    } else {
      next()
    }
  }
})

export default router
