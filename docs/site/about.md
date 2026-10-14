---
title: About fbrcm
description: What fbrcm does, why it requests Google access, and how it handles your data.
---

<script setup>
import { VPImage } from 'vitepress/theme'
</script>

# About fbrcm

fbrcm is an independent, open-source desktop terminal application for managing
Firebase Remote Config. Its terminal interface and command-line tools let you
discover projects, inspect and compare configuration, prepare local drafts,
validate changes, and publish or restore templates. It also reads registered
Firebase app information and downloads SDK configuration.

<VPImage
  :image="{
    light: '/fbrcm-light.png',
    dark: '/fbrcm-dark.png',
    alt: 'fbrcm terminal interface showing Firebase Remote Config projects and parameters'
  }"
/>

## Why fbrcm requests Google access

Google authorization lets fbrcm access the Google Cloud and Firebase projects
that your account is permitted to manage. It uses the Cloud Resource Manager,
Firebase Remote Config, and Firebase Management APIs for the operations you
request. Google Cloud IAM permissions control which projects and actions are
available to your account.

Google API requests run directly from your computer. Credentials, cached
configuration, and drafts remain on your device unless you explicitly export
them or configure other programs to handle them. fbrcm has no developer-operated
backend, telemetry, or advertising. The developer does not receive your Google
user data or use it to train AI or machine-learning models.

## Public documentation and privacy

This documentation website is public and requires no login. Google authorization
takes place in the installed application when accessing your projects.

Read the [Privacy policy](/privacy-policy) for details about Google data access,
storage, sharing, security, retention, and deletion, and the
[Terms of Service](/terms) for the terms of using fbrcm. fbrcm is not affiliated
with or endorsed by Google.

## Get started

Follow the [installation and setup guide](/guide/) to start using fbrcm, or
read about [authentication and project discovery](/guide/authentication) before
choosing a credential. The [source code and releases](https://github.com/yumauri/fbrcm)
are available on GitHub under the MIT License.
