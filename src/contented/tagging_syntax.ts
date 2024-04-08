import {ApiDef} from './api_def';
import {Injectable} from '@angular/core';
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

let results: Array<Tag> = [];
export let TAGS_RESPONSE = {
  total: -1,
  initialized: false,
  results: results,
};

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


      // MultiWord tags would need to have a different matcher (and remove the hack)
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
      [/'/, 'string.invalid'],

      // Matching wordlike bounds but this absorbs tokens and then the typeKeywords do not work
      [/[a-zA-Z_][\w$]*/, { 
        cases: {
         '@keywords': 'keyword',
         '@typeKeywords': 'type',
         '@operators': 'operator',
          } 
      }],
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



// First get a system that will build out and grab problem tags
// Grab the regex that this builds and have it useful in the vscode_editor
// Make it so the setMonacoLanguage is a class method that is smart enough to build out the hackery options
// Fix the unit test so that it can also be run with a load language call.
// Fix the API to have pages, search etc.

// Could make it so this takes the language name
@Injectable()
export class TagLang {

  constructor() {

  }

  createHackeryMatcher(tags: Array<string>): RegExp|undefined {
    let hackery: Array<string> = [];
    _.each(tags, tag => {
      let arr = tag ? tag.split(" ") : [];
      if (arr && arr.length > 1) {
        hackery.push(tag);
      }
    });
    if (!_.isEmpty(hackery)) {
      return new RegExp(hackery.join("|"));
    }
    return undefined
  }

  setMonacoLanguage(languageName: string, keywords: Array<string>, typeKeywords: Array<string>, operators: Array<string> = []) {
    let lang = (window as any).monaco.languages;
    let syntax = _.clone(TAGGING_SYNTAX);

    // HACKERY!   WEEEEE
    syntax.keywords = keywords || []
    syntax.typeKeywords = typeKeywords || [];
    syntax.operators = _.isEmpty(operators) ? syntax.operators : operators;
    lang.register({id: languageName, configuration: syntax})

    // Doesn't exactly work, there needs to be a unity between type keyword matching?
    // The same word pattern does NOT make the wordAtPosition API play nice with the token offset
    lang.setLanguageConfiguration(languageName, {
      //wordPattern: /'?\w[\w'-.]*[?!,;:"]*/
      wordPattern: /(-?\d*\.\d\w*)|([^\`\~\!\#\%\^\&\*\(\)\-\=\+\{\}\\\|\;\:\'\"\,\.\<\>\/\?\s]+)/g,
    });

    // These allow for matching specific keys that are multi-word or a pain to match.
    let keywordsHack = this.createHackeryMatcher(keywords);
    let typesHack = this.createHackeryMatcher(typeKeywords);
    let operatorsHack = this.createHackeryMatcher(operators);
    if (keywordsHack) {
      // console.log("Keywords in the hack", keywordsHack)
      syntax.tokenizer.root.push([keywordsHack, 'keyword']);
    }
    if (typesHack) {
      // console.log("Types hack", typesHack)
      syntax.tokenizer.root.unshift([typesHack, 'type']);
    }
    if (!_.isEmpty(operatorsHack)) {
      syntax.tokenizer.root.unshift([operatorsHack, 'number']);
    }
    lang.setMonarchTokensProvider(languageName, syntax);
    lang.registerCompletionItemProvider(languageName, {
      provideCompletionItems: (model, position) => {
        const suggestions = [
        ...this.getSuggestionsForType(lang.CompletionItemKind.Keyword, keywords),
        ...this.getSuggestionsForType(lang.CompletionItemKind.Type, typeKeywords),
        ...this.getSuggestionsForType(lang.CompletionItemKind.Number, operators),
        ];
        return { suggestions: suggestions }
      }
    });
    return syntax
  }

  // Would be nice to get these imported properly with typing
  // TODO: Should this only suggest lower case?
  getSuggestionsForType(kind: number, tags: Array<string>) {
    return tags.map((val: string, _idx) => {
      return {
        label: val,
        kind: kind,
        insertText: val?.toLowerCase(),
      }
    });
  }

  loadLanguage(monaco: any, languageName: string) {
    $.ajax(ApiDef.contented.tags, {
      params: {per_page: 1000},
      success: res => {
        // I should also change the color of the type and the keyword.
        let results = res.results;

        let tags = _.map(results, r => new Tag(r));
        let keywordTags = _.map(_.filter(tags, {tag_type: 'keywords'}), 'id');
        let keywords = keywordTags.concat(
          _.map(languages, lang => _.upperFirst(lang)),
          _.map(languages, lang => lang.toUpperCase())
        );
        let typeKeywordTags = _.map(_.filter(tags, {tag_type: 'typeKeywords'}), 'id');
        let operators = _.map(_.filter(tags, {tag_type: 'operators'}), 'id');
        
        this.setMonacoLanguage("tagging", keywords, typeKeywordTags, operators);

        // Load this data once.
        TAGS_RESPONSE.total = res.total;
        TAGS_RESPONSE.results = tags;
        TAGS_RESPONSE.initialized = true;
      }, error: err => {
        //this.setMonacoLanguage("tagging", [], []);
        console.error("loadLanguage failed to load tags", err)
      }
    });
  }
}
