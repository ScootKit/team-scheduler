<div align="center">

<img src="./.github/assets/images/logo.svg" width="200px" alt="Timeful logo" />

</div>
<br />
<div align="center">

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-orange.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Donate](https://img.shields.io/badge/-Donate%20with%20Paypal-blue?logo=paypal)](https://www.paypal.com/donate/?hosted_button_id=KWCH6LGJCP6E6)
[![X (formerly Twitter) Follow](https://img.shields.io/twitter/follow/timeful_app?label=%40timeful_app&labelColor=white)](https://x.com/timeful_app)
[![Discord](https://img.shields.io/badge/-Join%20Discord-7289DA?logo=discord&logoColor=white)](https://discord.gg/v6raNqYxx3)
[![Subreddit subscribers](https://img.shields.io/reddit/subreddit-subscribers/schej?label=join%20r%2Fschej)](https://www.reddit.com/r/schej/)

</div>

<img src="./.github/assets/images/hero.jpg" alt="Timeful hero" />

Timeful is a scheduling platform helps you find the best time for a group to meet. It is a free availability poll that
is easy to use and integrates with your calendar.

Hosted version of the site: https://timeful.app

Built
with [Vue 2](https://github.com/vuejs/vue), [MongoDB](https://github.com/mongodb/mongo), [Go](https://github.com/golang/go),
and [TailwindCSS](https://github.com/tailwindlabs/tailwindcss)

## ⚡ WannPassts fork (ScootKit internal)

This repository (`ScootKit/team-scheduler`) is an internal fork of upstream
[`schej-it/timeful.app`](https://github.com/schej-it/timeful.app), rebranded
**WannPassts** ("Wann passt's?"). It is a private, internal scheduling tool for
ScootKit — **not** a general-purpose product. It is intended for meetings with
at least one external participant; for internal employee-to-employee scheduling,
use Google Calendar's own tools.

It has diverged substantially from upstream. Notable changes:

**Access & accounts**

- **Domain-gated sign-in** via an email-domain allowlist (`ALLOWED_EMAIL_DOMAINS`,
  e.g. `scootkit.com,helper.scootkit.co`). Employees (`@scootkit.com`) sign in
  with Google OAuth or email OTP; helpers (`@helper.scootkit.co`) sign in with
  email **OTP delivered via Amazon SES**. A clear error is shown for disallowed
  domains.
- **Anonymous event creation disabled** (sign-in required); guests can still
  view/respond to shared event links, and must **record privacy-policy consent**
  (timestamp + policy version) before submitting.
- Last name is **optional**.

**Removed**

- **Monetization & paywalls** — Stripe, premium, and the free-event limit
  (everyone is treated as premium).
- **All ads & third-party tracking** — Publift, Primis, Carbon ads; Google Tag
  Manager; **PostHog (fully stripped — no calls to `e.timeful.app`)**; the
  `geolocation-db.com` IP lookup.
- **Public marketing surface** — landing page (→ redirects to sign-in), Reddit/
  GitHub nudges, the (broken) "Convert When2meet" tool, `contact@` + social links.
- **Email features** — reminder emails, "collect respondents' email", and
  join-notification emails.
- **Bots** — the Slack bot (it mounted an unauthenticated `/api/slackbot`
  endpoint) and the dead Discord bot are no longer wired up.
- **Calendar providers** — **Apple Calendar (CalDAV)** and **Outlook (Microsoft
  Graph)** support is removed, which also drops the `github.com/jonyTF/go-webdav`
  fork and the `@azure/msal-*` deps. **Only Google + ICS** calendar import remain.

**Added**

- **Response deadlines** — optional per-event cutoff with a live countdown,
  banner, and grid highlight; no submissions accepted after it.
- **Suggested topics** — respondents can suggest discussion topics/agenda items
  (per-event opt-in/out by the creator); highlighted right after they respond.
- **Scheduled event** — the creator sets the final date/time + an optional Google
  Meet link; shown as a banner + highlighted in the grid. Integrates with the
  existing "schedule event" flow (opens the dialog prefilled).
- **24-hour time** and **European day-first dates** (`22.6`) by default.

**Security hardening**

- Ported upstream's pending security PRs (session-cookie hardening, ICS SSRF
  protection, OTP send rate-limiting, analytics Basic-Auth fail-closed,
  calendar-parser & event-route hardening).
- Additional: Google `email_verified` enforced, `check-email` enumeration fixed,
  OTP delivered via SES with **no code logging in production**, `editEvent`
  requires auth, per-event topic flood cap, ICS feed size cap, meeting links
  restricted to `http(s)`. Bumped `gin-contrib/cors` to 1.6.0.

**Branding / legal**

- `WannPassts` wordmark throughout; **Impressum** + external **Privacy Policy**
  links (env-driven — see below); an AGPL **"Source code"** link in the footer.

Upstream identifiers (Go module `schej.it/server`, Mongo DB `schej-it`) are
intentionally left unchanged.

### Relationship to upstream

Upstream is kept as the `upstream` git remote **only** so we can occasionally
review and **cherry-pick individual security fixes**:

```bash
git fetch upstream
git log upstream/main   # review; cherry-pick only relevant security commits
```

We do **not** merge upstream wholesale — this fork has diverged too far
(removed providers/features, different access model), so clean merges aren't
expected. `main` is the canonical, deployed, AGPL-published version.

Like upstream, this fork is **AGPL-3.0**, and we're glad to share our changes
back as open source. Since WannPassts is reachable by external users over the
network, the complete modified source lives here and is linked from the app
footer, in the spirit of AGPL — thanks to the upstream maintainers for building
a great tool to start from.

### Support & self-hosting

> This repository is shared as open source, but it's tailored to ScootKit's
> internal setup (domain-gated sign-in, our legal pages, removed integrations,
> etc.) rather than as a general-purpose product — so we're **not able to offer
> support** for self-hosting or deploying it.
>
> If you'd like a scheduling tool you can run yourself, we'd genuinely recommend
> the original, actively-maintained upstream project:
> **[schej-it/timeful.app](https://github.com/schej-it/timeful.app)** (or the
> hosted version at https://timeful.app) — it's the better starting point, and
> self-hosting questions are best directed there.

### Configuration (fork-specific env)

These env vars control fork behavior (see `server/.env.template` and
`compose.yaml`):

| Variable                                                                    | Purpose                                                                                                             |
|-----------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `ALLOWED_EMAIL_DOMAINS`                                                     | Comma-separated sign-in domain allowlist (empty = allow all)                                                        |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM` | Email transport (Amazon SES) for OTP + system mail                                                                  |
| `FRONTEND_BASE_URL`                                                         | Public base URL used in OG tags / email links                                                                       |
| `VUE_APP_PRIVACY_POLICY_URL` / `VUE_APP_IMPRINT_URL`                        | Legal links shown in the footer (build-time; **unset = hidden**, so a mirror never surfaces ScootKit's legal pages) |

## Demo

[![demo video](http://markdown-videos-api.jorgenkh.no/youtube/vFkBC8BrkOk)](https://www.youtube.com/watch?v=vFkBC8BrkOk)

## Features

- See when everybody's availability overlaps
- Easily specify date + time ranges to meet between
- Google calendar, Outlook, Apple calendar integration
- "Available" vs. "If needed" times
- Determine when a subset of people are available
- Schedule across different time zones
- Email notifications + reminders
- Duplicating polls
- Availability groups - stay up to date with people's real-time calendar availability
- Export availability as CSV
- Only show responses to event creator

## Plugin API

Read these docs to design your own browser plugins to get + set availability on Timeful events programmatically!

[Plugin API Docs](./PLUGIN_API_README.md)

## Self-hosting

See the [Deployment Guide](./DEPLOYMENT.md) for Docker Compose and NixOS setup instructions.
