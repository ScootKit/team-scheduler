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

Timeful is a scheduling platform helps you find the best time for a group to meet. It is a free availability poll that is easy to use and integrates with your calendar.

Hosted version of the site: https://timeful.app

Built with [Vue 2](https://github.com/vuejs/vue), [MongoDB](https://github.com/mongodb/mongo), [Go](https://github.com/golang/go), and [TailwindCSS](https://github.com/tailwindlabs/tailwindcss)

## ⚡ WannPassts fork (ScootKit internal)

This repository (`ScootKit/team-scheduler`) is an internal fork of upstream
[`schej-it/timeful.app`](https://github.com/schej-it/timeful.app), rebranded
**WannPassts** ("Wann passt's?"). It is a private scheduling tool for ScootKit,
not a public product. Key differences from upstream:

- **Internal access only** — sign-in is restricted to an email-domain allowlist
  (`ALLOWED_EMAIL_DOMAINS`, e.g. `scootkit.com,helper.scootkit.co`). Employees
  use Google OAuth or email OTP; helpers use email OTP (delivered via Amazon
  **SES** — `SMTP_*` env vars). Anonymous event **creation** is disabled
  (sign-in required); guests can still view/respond to shared event links.
- **No monetization or tracking** — Stripe/premium and the free-event limit are
  removed (everyone is treated as premium), and all ad/analytics scripts
  (Publift, Google Tag Manager, Primis) are stripped.
- **No public marketing surface** — the landing page is removed; `/` redirects
  to sign-in (authenticated users go to their dashboard).
- **DE/EU legal** — footer links to an external Imprint
  (`scootkit.com/impressum`) and Privacy Policy (`scootk.it/scnx-privacy`);
  guests must record privacy-policy consent before submitting availability.
- **Added features** — optional per-event **response deadlines** (no
  submissions after the cutoff).
- **Security** — upstream's pending security fixes are ported in (session-cookie
  hardening, ICS SSRF protection, OTP send rate-limiting, analytics auth
  hardening, calendar-parser and event-route hardening).
- Upstream identifiers (Go module `schej.it/server`, Mongo DB `schej-it`) are
  intentionally left unchanged.

### Staying in sync with upstream

Upstream is tracked as the `upstream` git remote. To pull in new upstream
security fixes:

```bash
git fetch upstream
# review and cherry-pick / merge the relevant commits onto this branch
```

Upstream remains **AGPL-3.0**. This fork is operated as a network service that
external users can access, so under **AGPL §13** we publish our complete
modified source here and link to it from the app footer.

### Support & self-hosting

> **We do not provide any support for self-hosting, deploying, or otherwise
> running this fork.** It is published solely to meet the AGPL source-availability
> obligation and is tailored to ScootKit's internal needs (domain-gated sign-in,
> our branding, our legal pages, etc.). It is **not** a general-purpose product.
>
> If you want a scheduling tool you can run yourself, please use the original,
> actively-maintained upstream project instead:
> **[schej-it/timeful.app](https://github.com/schej-it/timeful.app)** (or the
> hosted version at https://timeful.app). Direct self-hosting questions there,
> not here.

### Configuration (fork-specific env)

These env vars control fork behavior (see `server/.env.template` and
`compose.yaml`):

| Variable | Purpose |
| --- | --- |
| `ALLOWED_EMAIL_DOMAINS` | Comma-separated sign-in domain allowlist (empty = allow all) |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM` | Email transport (Amazon SES) for OTP + system mail |
| `FRONTEND_BASE_URL` | Public base URL used in OG tags / email links |
| `VUE_APP_PRIVACY_POLICY_URL` / `VUE_APP_IMPRINT_URL` | Legal links shown in the footer (build-time; **unset = hidden**, so a mirror never surfaces ScootKit's legal pages) |

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
