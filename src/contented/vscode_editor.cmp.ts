/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
import { Component, OnInit, Input, Output, EventEmitter, ViewChild, ElementRef} from '@angular/core';
import {ActivatedRoute } from '@angular/router';
import {FormBuilder, FormControl, FormGroup } from '@angular/forms';
import {debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {EditorComponent} from 'ngx-monaco-editor-v2';
import {ContentedService} from './contented_service';
import { GlobalBroadcast } from './global_message';

import * as _ from 'lodash-es';
import { Tag } from './content';

interface VSCodeChange {
  value: string;
  tags: Array<string>;
}

@Component({
  selector: 'vscode-editor-cmp',
  templateUrl: './vscode_editor.ng.html',
})
export class VSCodeEditorCmp implements OnInit {

  @ViewChild('vseditor') editor?: EditorComponent;
  @ViewChild('container') container?: ElementRef<HTMLDivElement>;

  @Input() editForm?: FormGroup;
  @Input() editorValue: string = "";
  @Input() descriptionControl?: FormControl<string>;
  @Input() showTagging: boolean = false;
  @Input() readOnly: boolean = true;
  @Input() language: string = "tagging";

  @Input() tags: Array<Tag> = [];
  @Input() editorOptions = {
    theme: 'vs-dark',
    language: this.language,
  };
  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  @Output() changeEmitter = new EventEmitter<VSCodeChange>();


  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: any;
  public initialized = false;
  public problemTags: Array<Tag> = [];
  public tagLookup: {[id: string]: Tag} = {};

  constructor(public fb: FormBuilder, public route: ActivatedRoute, public _service: ContentedService) {
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    this.editorOptions.language = this.language;

    if (!this.editForm) {
      this.editForm = this.fb.group({});
    }
    let control = this.descriptionControl || this.editForm.get("description") || new FormControl(this.editorValue || "");
    this.editForm.addControl("description", control);
    this.editorValue = this.editorValue || control.value;
    this.descriptionControl = (control as FormControl<string>);  // Sketchy...

    this.monacoDigestHackery();
    this.tags?.length > 0 ? this.assignTagLookup(this.tags) : this.loadTags();
  }

  loadTags() {
    this._service.getTags().subscribe({
      next: (tagRes: {total: number, results: Tag[]}) => {
        this.assignTagLookup(tagRes.results || []);
      }, 
      error: err => { 
        GlobalBroadcast.error('Failed to load tags', err);
      }
    });
  }

  assignTagLookup(tags: Array<Tag>) {
    this.tags = tags;
    this.tagLookup = _.keyBy(this.tags, 'id');
    this.problemTags = _.filter(this.tags, t => {
      return t.isProblem();
    });
  }

  // The onInit from monaco pulls us OUT of a proper digest detection, so if I set initialized
  // directly in the 'afterMonacoInit' it is not detected till an edit.  This gets us insight into
  // if the monacoEditor is present and fixes any digest loop redraw errors...
  monacoDigestHackery(count: number = 0) {
    _.delay(() => {
      if (this.monacoEditor || count > 4) {
        this.initialized = true;  // Eventually we want to give up.. probably the editor bailed.
      } else {
        this.monacoDigestHackery(count + 1);
      }
    }, 500);
  }

  setReadOnly(state: boolean) {
    this.readOnly = state;
    if (this.monacoEditor) {
      this.monacoEditor.updateOptions({readOnly: this.readOnly});
    }
    if (this.editForm) {
      if (this.readOnly) {
        this.editForm.disable();
      } else {
        this.editForm.enable();
      }
    }
  }

  // The pure Monaco part is definitely worth an indepenent component (I think)
  afterMonacoInit(monaco: any) {
    this.monacoEditor = monaco;
    (window as any).M = monaco;
      // This is a little awkward but we need to be able to change the form control
    if (this.editor) {
      this.changeEmitter.pipe(
        distinctUntilChanged(),
        debounceTime(10)
      ).subscribe({
        next: (evt: VSCodeChange) => {
          console.log("Debounce time", evt)
            this.editForm.get("description").setValue(evt.value);
        },
        error: (err) => {
          GlobalBroadcast.error('Monaco init failed', err);
        }
      });

      this.editor.registerOnChange((val: string) => {
        this.changeEmitter.emit({
          tags: this.getTokens(),
          value: val
        });
      });
    }
    this.afterMonaco();
  }

  // TokenType match must be set smarter.
  public getTokens(tokenType: string = "keyword", language = this.language) {
    // Dynamically loaded
    let monaco = (window as any).monaco;
    let tags = new Set<string>()
    if (monaco && this.descriptionControl) {
      let tokenArr = monaco.editor.tokenize(this.descriptionControl.value, language);
      let m = this.monacoEditor.getModel();

      // Custom tag matching per line I guess where we remove the matched tags from the loop but it might
      // actually be faster to just do a contains against the entire string.
      // The word offset boundry is all messed up...
      let currentLang = _.find(monaco.languages.getLanguages(), {id: language});
      // console.log("currentLanguage", currentLang, monaco.languages.getLanguages());

      let match = `${tokenType}.${language}`;
      _.each(tokenArr, (tokens, lineIdx) => {
        let line = m.getLineContent(lineIdx + 1)
        _.each(tokens, token => {
          //console.log("token.type", token.type)
          if (token.type == match) {
            // This should work, but doesn't because of crazy word boundry monaco stuff?
            // let position = m.getPositionAt(token.offset + 1);
            // position.lineNumber = lineIdx + 1;
            // let word = m.getWordAtPosition(position);

            // The highlights work but the word positions are all jacked up
            let word = this.readToken(line, token.offset)
            if (this.tagLookup[word]) {
              tags.add(word);
            } else {
              this.processProblemTags(line, tags);
            }
          }
        })
      });
    }
    console.log('tags', tags);
    return Array.from(tags);
  }

  processProblemTags(line: string, tags: Set<string>) {
    // console.log("Problem line", line);
    _.each(this.problemTags, t => {
      if (line.includes(t.id)) {
        tags.add(t.id)
      }
    });
  }

  // Additional keys for token matching?
  readToken(str: string, offset: number) {
    let code, j;
    let len = str.length;
    if (offset > len) {
      return "";
    }
    for (j = offset; j < len; ++j) {
      code = str.charCodeAt(j);
      if (!(code > 47 && code < 58) && // numeric (0-9)
          !(code > 64 && code < 91) && // upper alpha (A-Z)
          !(code === 95) && // upper alpha (A-Z)
          !(code > 96 && code < 123)) { // lower alpha (a-z)
            break;
      }
    }
    return str.slice(offset, j)
  };

  public afterMonaco() {
    console.log("After monaco initialization.");
    if (!this.editForm) {
      return;
    }

    // Subscribes specifically to the description changes.
    let control = this.editForm.get("description");
    if (control) {
      control.valueChanges.pipe(
        distinctUntilChanged(),
        debounceTime(500)
      ).subscribe({
        next: (evt: any) => {
          if (this.editForm) {
            // console.log("VS Code editor form updated", this.editForm.value);
          }
        }
      });
      // Set this after the initialization.
      this.editorValue = control.value;
    }
    _.delay(() => {
      this.fitContent();
    }, 500);
  }


  // TODO:  This also needs to handle a window resize event to actually check the content 
  // and do a redraw.
  fitContent() {
    let el = this.container.nativeElement;
    let width = el.offsetWidth;

    let updateHeight = () => {
      let editor = this.monacoEditor;
      const lineCount = Math.max(editor.getModel()?.getLineCount(), 8);

      // You would think this would work but unfortunately the height of content is altered
      // by the spacing of the render so it expands forever.
      //const contentHeight = Math.min(2000, this.monacoEditor.getContentHeight());

      let contentHeight = 19 * (lineCount + 2);
      el.style.width = `${width}px`;
      el.style.height = `${contentHeight}px`;
      editor.layout({width, height: contentHeight });
    };

    _.delay(() => {
      updateHeight();
    }, 100);
    let changed = _.debounce(updateHeight, 150);
    this.monacoEditor.onDidChangeModelDecorations(changed);
  }
}