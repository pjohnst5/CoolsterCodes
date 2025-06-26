"use strict";
// Licensed to the .NET Foundation under one or more agreements.
// The .NET Foundation licenses this file to you under the MIT license.
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
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
exports.options = options;
exports.meta = meta;
exports.loc = loc;
exports.breakWord = breakWord;
exports.breakWordLit = breakWordLit;
exports.isExternalHref = isExternalHref;
exports.isSameURL = isSameURL;
const lit_html_1 = require("lit-html");
function options() {
    return __awaiter(this, void 0, void 0, function* () {
        return yield Promise.resolve().then(() => __importStar(require('./main.js'))).then(m => m.default);
    });
}
/**
 * Get the value of an HTML meta tag.
 */
function meta(name) {
    var _a;
    return (_a = document.querySelector(`meta[name="${name}"]`)) === null || _a === void 0 ? void 0 : _a.content;
}
/**
 * Gets the localized text.
 * @param id key in token.json
 * @param args arguments to replace in the localized text
 */
function loc(id, args) {
    let result = meta(`loc:${id}`) || id;
    if (args) {
        for (const key in args) {
            result = result.replace(`{${key}}`, args[key]);
        }
    }
    return result;
}
/**
 * Add <wbr> into long word.
 */
function breakWord(text) {
    if (!text) {
        return [];
    }
    const regex = /([a-z0-9])([A-Z]+[a-z])|([a-zA-Z0-9][.,/<>_])/g;
    const result = [];
    let start = 0;
    while (true) {
        const match = regex.exec(text);
        if (!match) {
            break;
        }
        const index = match.index + (match[1] || match[3]).length;
        result.push(text.slice(start, index));
        start = index;
    }
    if (start < text.length) {
        result.push(text.slice(start));
    }
    return result;
}
/**
 * Add <wbr> into long word.
 */
function breakWordLit(text) {
    const result = [];
    breakWord(text).forEach(word => {
        if (result.length > 0) {
            result.push((0, lit_html_1.html) `<wbr>`);
        }
        result.push((0, lit_html_1.html) `${word}`);
    });
    return (0, lit_html_1.html) `${result}`;
}
/**
 * Check if the url is external.
 * @param url The url to check.
 * @returns True if the url is external.
 */
function isExternalHref(url) {
    return url.hostname !== window.location.hostname || url.protocol !== window.location.protocol;
}
/**
 * Determines if two URLs should be considered the same.
 */
function isSameURL(a, b) {
    return normalizeUrlPath(a) === normalizeUrlPath(b);
    function normalizeUrlPath(url) {
        return url.pathname
            .replace(/\/index\.html$/gi, '/')
            .replace(/\.html$/gi, '')
            .replace(/\/$/gi, '')
            .toLowerCase();
    }
}
