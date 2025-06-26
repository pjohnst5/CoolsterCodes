// Licensed to the .NET Foundation under one or more agreements.
// The .NET Foundation licenses this file to you under the MIT license.

import 'bootstrap'
import { options } from './helper'
import { enableSearch } from './search'

import 'bootstrap-icons/font/bootstrap-icons.scss'
import './docfx.scss'

declare global {
  interface Window {
    docfx: {
      ready?: boolean,
      searchReady?: boolean,
      searchResultReady?: boolean,
    }
  }
}

async function init() {
  window.docfx = window.docfx || {}

  const { start } = await options()
  start?.()

  await Promise.all([
    enableSearch()
  ])

  window.docfx.ready = true
}

init().catch(console.error)
