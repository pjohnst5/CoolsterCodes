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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const lunr_1 = __importDefault(require("lunr"));
const lunr_stemmer_support_1 = __importDefault(require("lunr-languages/lunr.stemmer.support"));
const tinyseg_1 = __importDefault(require("lunr-languages/tinyseg"));
const lunr_multi_1 = __importDefault(require("lunr-languages/lunr.multi"));
const idb_keyval_1 = require("idb-keyval");
let search;
function loadIndex(_a) {
    return __awaiter(this, arguments, void 0, function* ({ lunrLanguages }) {
        const { index, data } = yield loadIndexCore();
        search = q => index.search(q).map(({ ref }) => data[ref]);
        postMessage({ e: 'index-ready' });
        function loadIndexCore() {
            return __awaiter(this, void 0, void 0, function* () {
                const res = yield fetch('index.json');
                const etag = res.headers.get('etag');
                const data = yield res.json();
                const cache = (0, idb_keyval_1.createStore)('docfx', 'lunr');
                if (lunrLanguages && lunrLanguages.length > 0) {
                    (0, lunr_multi_1.default)(lunr_1.default);
                    (0, lunr_stemmer_support_1.default)(lunr_1.default);
                    if (lunrLanguages.includes('ja')) {
                        (0, tinyseg_1.default)(lunr_1.default);
                    }
                    yield Promise.all(lunrLanguages.map(initLanguage));
                }
                if (etag) {
                    const value = JSON.parse((yield (0, idb_keyval_1.get)('index', cache)) || '{}');
                    if (value && value.etag === etag) {
                        return { index: lunr_1.default.Index.load(value), data };
                    }
                }
                const index = (0, lunr_1.default)(function () {
                    lunr_1.default.tokenizer.separator = /[\s\-.()]+/;
                    this.ref('href');
                    this.field('title', { boost: 50 });
                    this.field('keywords', { boost: 40 });
                    this.field('summary', { boost: 20 });
                    if (lunrLanguages && lunrLanguages.length > 0) {
                        this.use(lunr_1.default.multiLanguage(...lunrLanguages));
                    }
                    for (const key in data) {
                        this.add(data[key]);
                    }
                });
                if (etag) {
                    yield (0, idb_keyval_1.set)('index', JSON.stringify(Object.assign(index.toJSON(), { etag })), cache);
                }
                return { index, data };
            });
        }
    });
}
onmessage = function (e) {
    if (e.data.q && search) {
        postMessage({ e: 'query-ready', d: search(e.data.q) });
    }
    else if (e.data.init) {
        loadIndex(e.data.init).catch(console.error);
    }
};
const langMap = {
    ar: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ar.js'))),
    da: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.da.js'))),
    de: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.de.js'))),
    du: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.du.js'))),
    el: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.el.js'))),
    es: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.es.js'))),
    fi: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.fi.js'))),
    fr: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.fr.js'))),
    he: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.he.js'))),
    hi: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.hi.js'))),
    hu: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.hu.js'))),
    hy: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.hy.js'))),
    it: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.it.js'))),
    ja: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ja.js'))),
    jp: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.jp.js'))),
    kn: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.kn.js'))),
    ko: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ko.js'))),
    nl: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.nl.js'))),
    no: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.no.js'))),
    pt: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.pt.js'))),
    ro: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ro.js'))),
    ru: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ru.js'))),
    sa: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.sa.js'))),
    sv: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.sv.js'))),
    ta: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.ta.js'))),
    te: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.te.js'))),
    th: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.th.js'))),
    tr: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.tr.js'))),
    vi: () => Promise.resolve().then(() => __importStar(require('lunr-languages/lunr.vi.js')))
    // zh is currently not supported due to dependency on NodeJS.
    // zh: () => import('lunr-languages/lunr.zh.js')
};
function initLanguage(lang) {
    return __awaiter(this, void 0, void 0, function* () {
        if (lang !== 'en') {
            const { default: init } = yield langMap[lang]();
            init(lunr_1.default);
        }
    });
}
