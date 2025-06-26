"use strict";
// Licensed to the .NET Foundation under one or more agreements.
// The .NET Foundation licenses this file to you under the MIT license.
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
Object.defineProperty(exports, "__esModule", { value: true });
require("bootstrap");
const helper_1 = require("./helper");
const search_1 = require("./search");
require("bootstrap-icons/font/bootstrap-icons.scss");
require("./docfx.scss");
function init() {
    return __awaiter(this, void 0, void 0, function* () {
        window.docfx = window.docfx || {};
        const { start } = yield (0, helper_1.options)();
        start === null || start === void 0 ? void 0 : start();
        yield Promise.all([
            (0, search_1.enableSearch)()
        ]);
        window.docfx.ready = true;
    });
}
init().catch(console.error);
