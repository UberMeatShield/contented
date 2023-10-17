import {ApiDef} from './api_def';
import {Injectable} from '@angular/core';
import {ContentedService} from './contented_service';
import {Tag} from './content';


import * as $ from 'jquery';
import * as _ from 'lodash-es';

let languages = [];

let technologies = [];
let operators = [
    '=', '>', '<', '!', '~', '?', ':', '==', '<=', '>=', '!=',
    '&&', '||', '++', '--', '+', '-', '*', '/', '&', '|', '^', '%',
    '<<', '>>', '>>>', '+=', '-=', '*=', '/=', '&=', '|=', '^=',
    '%=', '<<=', '>>=', '>>>='
  ];

// Restrictive email format so the highlights do not fight with other elements (no UC)
let mailFormat = /^[a-z0-9.!$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;
//let mailFormat = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;

// Mostly empty tagging support, dynamically load the tags and THEN register the language.
// https://stackoverflow.com/questions/52700307/how-to-use-monaco-editor-for-syntax-highlighting
export let TAGGING_SYNTAX = {
  // Set defaultToken to invalid to see what you do not tokenize yet
  // defaultToken: 'invalid',

  // These should be loaded from the API
  keywords: languages,
  typeKeywords: technologies,
  operators: operators,

  // we include these common regular expressions
  symbols:  /[=><!~?:&|+\-*\/\^%]+/,

  // C# style strings
  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

  // Complex tokenizer example (awkward to use with other matchers)
  tokenizer: {
    root: [
      // to show sections names nicely
      [mailFormat, 'type.identifier'],
      [/^[A-Z].*\./, 'type.identifier'], 


      // Matching wordlike bounds but this absorbs tokens and then the typeKeywords do not work
      [/[a-zA-Z_][\w$]*/, { 
        cases: {
         '@typeKeywords': 'keyword',
         '@keywords': 'keyword',
          } 
      }],

      // MultiWord tags would need to have a different matcher (and remove the hack)
      [/\w+\s\w+/, { 
        cases: {
         '@typeKeywords': 'typeKeyword',
         //'@keywords': 'keyword',
          } 
      }],

      // whitespace
      { include: '@whitespace' },

      // delimiters and operators
      [/[{}()\[\]]/, '@brackets'],
      [/[<>](?!@symbols)/, '@brackets'],
      //[/@symbols/, { cases: { '@operators': 'operator',
      //                        '@default'  : '' } } ],

      // @ annotations.
      // As an example, we emit a debugging log message on these tokens.
      // Note: message are supressed during the first load -- change some lines to see them.
      [/  @\s*[a-zA-Z_\$][\w\$]*/, { token: 'annotation', log: 'annotation token: $0' }],

      // numbers
      [/\d*\.\d+([eE][\-+]?\d+)?/, 'number.float'],
      [/0[xX][0-9a-fA-F]+/, 'number.hex'],
      [/\d+/, 'number'],

      // delimiter: after number because of .\d floats
      [/[;,.]/, 'delimiter'],

      // strings
      [/"([^"\\]|\\.)*$/, 'string.invalid' ],  // non-teminated string
      [/"/,  { token: 'string.quote', bracket: '@open', next: '@string' } ],

      // characters
      [/'[^\\']'/, 'string'],
      [/(')(@escapes)(')/, ['string','string.escape','string']],
      [/'/, 'string.invalid']
    ],
    //wordPattern: /'?\w[\w'-.]*[?!,;:"]*/,

    // Whitespace comment is handling # comments
    comment: [
      [/[^\/*]+/, 'comment' ],
      [/\/\*/,    'comment', '@push' ],    // nested comment
      ["\\*/",    'comment', '@pop'  ],
      [/[\/*]/,   'comment' ],
    ],

    string: [
      [/[^\\"]+/,  'string'],
      [/@escapes/, 'string.escape'],
      [/\\./,      'string.escape.invalid'],
      [/"/,        { token: 'string.quote', bracket: '@close', next: '@pop' } ]
    ],

    whitespace: [
      [/[ \t\r\n]+/, 'white'],
      [/\s*(^#\s.*$)/, 'comment'],
      [/\/\*/,       'comment', '@comment' ],
      //[/\/\/.*$/,    'comment'],  Highlights links
    ],
  },
};


export function setMonacoLanguage(languageName: string, keywords: Array<string>, typeKeywords: Array<string>, operators: Array<string> = []) {
  let lang = (window as any).monaco.languages;
  let syntax = _.clone(TAGGING_SYNTAX);
  syntax.keywords = keywords || []
  syntax.typeKeywords = typeKeywords || [];
  syntax.operators = _.isEmpty(operators) ? syntax.operators : operators;
  lang.register({id: languageName, configuration: syntax})
  lang.setMonarchTokensProvider(languageName, syntax);

  // Doesn't exactly work, there needs to be a unity between type keyword matching?
  // The same word pattern does NOT make the wordAtPosition API play nice with the token offset
  lang.setLanguageConfiguration(languageName, {
    //wordPattern: /'?\w[\w'-.]*[?!,;:"]*/
    wordPattern: /(-?\d*\.\d\w*)|([^\`\~\!\#\%\^\&\*\(\)\-\=\+\{\}\\\|\;\:\'\"\,\.\<\>\/\?\s]+)/g,
  });
}

@Injectable()
export class TagLang {

  constructor() {

  }

  loadLanguage(monaco: any, languageName: string) {
    $.ajax(ApiDef.contented.tags, {
      success: res => {
        // I should also change the color of the type and the keyword.

        let tags = _.map(res, r => new Tag(r));

        let keywordTags = _.map(_.filter(tags, {tag_type: 'keywords'}), 'id');
        let keywords = keywordTags.concat(
          _.map(languages, lang => _.upperFirst(lang)),
          _.map(languages, lang => lang.toUpperCase())
        );
        let typeKeywordTags = _.map(_.filter(tags, {tag_type: 'typeKeywords'}), 'id');
        let operators = _.map(_.filter(tags, {tag_type: 'operators'}), 'id');
        setMonacoLanguage("tagging", keywords, typeKeywordTags, operators);
      }, error: err => {
        setMonacoLanguage("tagging", [], []);
        console.error("Failed to load tags", err)
      }
    });
    //console.log("Now here is where we register a new language for tags.");
    //let monaco = (<any>window).monaco;
  }
}
