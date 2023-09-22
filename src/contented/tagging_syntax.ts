import {ApiDef} from './api_def';
import {Injectable} from '@angular/core';
import {ContentedService} from './contented_service';


import * as $ from 'jquery';
import * as _ from 'lodash-es';

let languages = [
    'c#',   // A language that decides to use a common comment token.  Nice?
    'python',
    'typescript',
    'javascript', 
    'JavaScript', 
    'ruby',
    'perl',
    'Go',  // If you want to highlight too many things name a lang 'Go'
    'GoLang',
    'php',
    'java',
    'css', 
    'html'
];

let technologies = [
    'azure', 
    'django',
    'gobuffalo',
    'GoBuffalo',
    'flask',
    'bootstrap',
    'd3',
    'jira',
    'aws',
    'terraform',
    'gitlab',
    'ci',
    'GitLab',
    'ansible',
    'postgres',
    'mysql',
    'MySQL',
    'Oracle',
    'apache',
    'nginx',
    'rails',
    'iis',
    'EC2',
    'RDS',
    'S3', 
    'SQS',
    'jquery',
    'Google Earth',
    'Google Maps',
    'postgis',
    'Route53',
    'OpenSearch',
    'Angular'
];

// Restrictive email format so the highlights do not fight with other elements (no UC)
let mailFormat = /^[a-z0-9.!$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;
//let mailFormat = /^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$/;

// TODO: This will need to work a bit better
let langs = languages.concat(
  _.map(languages, lang => _.upperFirst(lang)),
  _.map(languages, lang => lang.toUpperCase())
 );
let techs = technologies.concat(
    _.map(technologies, tech => tech.toUpperCase()),
    _.map(technologies, tech => _.upperFirst(tech))
);

// Mostly empty tagging support, dynamically load the tags and THEN register the language.
// https://stackoverflow.com/questions/52700307/how-to-use-monaco-editor-for-syntax-highlighting
export let TAGGING_SYNTAX = {
  // Set defaultToken to invalid to see what you do not tokenize yet
  // defaultToken: 'invalid',

  // These should be loaded from the API
  keywords: langs,
  typeKeywords: techs,
  operators: [
    '=', '>', '<', '!', '~', '?', ':', '==', '<=', '>=', '!=',
    '&&', '||', '++', '--', '+', '-', '*', '/', '&', '|', '^', '%',
    '<<', '>>', '>>>', '+=', '-=', '*=', '/=', '&=', '|=', '^=',
    '%=', '<<=', '>>=', '>>>='
  ],

  // we include these common regular expressions
  symbols:  /[=><!~?:&|+\-*\/\^%]+/,

  // C# style strings
  escapes: /\\(?:[abfnrtv\\"']|x[0-9A-Fa-f]{1,4}|u[0-9A-Fa-f]{4}|U[0-9A-Fa-f]{8})/,

  // Complex tokenizer example
  tokenizer: {
    root: [
      // to show sections names nicely
      [mailFormat, 'type.identifier'],
      [/^[A-Z].*\./, 'type.identifier'], 
      [/C#|[a-zA-Z_$][\w$]*/, { 
        cases: {
         '@typeKeywords': 'keyword',
         '@keywords': 'keyword',
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

@Injectable()
export class TagLang {

  constructor() {
  }

  // Create something that will 

  loadLanguage(monaco: any, languageName: string) {
    $.ajax(ApiDef.contented.tags, {
      success: res => {
        let lang = monaco.languages;
        lang.register({id: languageName});
        lang.setMonarchTokensProvider(languageName, TAGGING_SYNTAX);
      }, error: err => {
        console.error("Failed to load up tags for the system");    
        let lang = monaco.languages;
        lang.register({id: languageName});
        lang.setMonarchTokensProvider(languageName, TAGGING_SYNTAX);

        // Have to add the Tags resource endpoint (does not exist)
      }
    });
    //console.log("Now here is where we register a new language for tags.");
    //let monaco = (<any>window).monaco;
  }
}
