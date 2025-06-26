// Licensed to the .NET Foundation under one or more agreements.
// The .NET Foundation licenses this file to you under the MIT license.

/**
 * Enables customization of the website through the global `window.docfx` object.
 */
export type DocfxOptions = {

  /** A list of [lunr languages](https://github.com/MihaiValentin/lunr-languages#readme) such as fr, es for full text search */
  lunrLanguages?: string[],

  /** Hooks to app start event */
  start?: () => void,
}
