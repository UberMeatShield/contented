/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
import { Component, OnInit, Input, Output, EventEmitter, ViewChild, ElementRef } from '@angular/core';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { EditorComponent } from 'ngx-monaco-editor-v2';
import { ContentedService } from './contented_service';
import { GlobalBroadcast } from './global_message';

// Why is it importing api.d?  Because Monaco does a bunch of css importing in the
// javascript which breaks the hell out of angular tooling, so just get the 'shapes'
// correct when doing a compile and move along.
import { KeyCode, editor as MonacoEditor } from 'monaco-editor/esm/vs/editor/editor.api.d';

import * as _ from 'lodash-es';
import { Tag, VSCodeChange } from './content';

@Component({
    selector: 'vscode-editor-cmp',
    templateUrl: './vscode_editor.ng.html',
    standalone: false
})
export class VSCodeEditorCmp implements OnInit {
  // The mix of actual M$ monaco types and the ngx-monaco-editor-v2 is a little hard to
  // grok. The M$ types are nice to have in for complex things but the naming gets confusing.
  @ViewChild('vseditor') editor?: EditorComponent;
  @ViewChild('container') container?: ElementRef<HTMLDivElement>;

  @Input() editForm?: FormGroup;
  @Input() editorValue: string = '';
  @Input() descriptionControl?: FormControl<string>;
  @Input() showTagging: boolean = false;
  @Input() readOnly: boolean = true;
  @Input() language: string = 'tagging';
  @Input() fixedLineCount: number = -1;
  @Input() placeholder: string;
  @Input() padLine: number = 1;

  @Input() tags: Array<Tag> = [];
  @Input() editorOptions = {
    theme: 'vs-dark',
    language: this.language,
  };

  // These are values for the Monaco Editors, change events are passed down int the form event
  // via the AfterInit. Then it can be read by other applications (after init should broadcast?).
  @Output() changeEmitter = new EventEmitter<VSCodeChange>();

  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: MonacoEditor.IStandaloneCodeEditor; // StandAlone?
  public initialized = false;
  public problemTags: Array<Tag> = [];
  public tagLookup: { [id: string]: Tag } = {};

  constructor(
    public fb: FormBuilder,
    public _service: ContentedService
  ) {}

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    this.tags?.length > 0 ? this.assignTagLookup(this.tags) : this.loadTags();
    this.editorOptions.language = this.language || this.editorOptions.language;

    if (!this.editForm) {
      this.editForm = this.fb.group({});
    }
    let control =
      this.descriptionControl || this.editForm.get('description') || new FormControl(this.editorValue || '');
    this.editForm.addControl('description', control);
    this.editorValue = this.editorValue || control.value;
    this.descriptionControl = control as FormControl<string>; // Sketchy...

    this.monacoDigestHackery();
  }

  loadTags() {
    this._service.getTags().subscribe({
      next: (tagRes: { total: number; results: Tag[] }) => {
        this.assignTagLookup(tagRes.results || []);
      },
      error: err => {
        GlobalBroadcast.error('Failed to load tags', err);
      },
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
        this.initialized = true; // Eventually we want to give up.. probably the editor bailed.
      } else {
        this.monacoDigestHackery(count + 1);
      }
    }, 500);
  }

  isInitialized() {
    let self = this;
    function monacoPoller(resolve, reject) {
      if (self.initialized) {
        resolve(self.initialized);
      } else {
        setTimeout(() => monacoPoller(resolve, reject), 500);
      }
    }

    return new Promise((resolve, reject) => {
      monacoPoller(resolve, reject);
    });
  }

  setReadOnly(state: boolean) {
    this.readOnly = state;
    if (this.monacoEditor) {
      this.monacoEditor.updateOptions({ readOnly: this.readOnly });
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
  afterMonacoInit(editorInstance: MonacoEditor.IStandaloneCodeEditor) {
    this.monacoEditor = editorInstance;

    // This is a little awkward but we need to be able to change the form control
    if (this.editor) {
      this.changeEmitter.pipe(distinctUntilChanged(), debounceTime(10)).subscribe({
        next: (evt: VSCodeChange) => {
          this.editForm.get('description').setValue(evt.value);
        },
        error: err => {
          GlobalBroadcast.error('Monaco init failed', err);
        },
      });

      // Probably need a post render or other change modifier manually if there is a value
      this.editor.registerOnChange((val: string) => {
        // This is awkward... the on change can call BEFORE the token / model is
        // fully updated so the tokens do not get updated correctly.
        _.delay(() => {
          this.changeEmitter.emit({
            tags: this.getTokens(),
            value: val,
          });
        }, 20);
      });

      // Allow Escape to unfocus. KeyCode.Escape === 9, the typing import
      // is not defined when actually running the code itself as it is just
      // type definition. Importing actual code == css compilation hell
      this.monacoEditor.addCommand(KeyCode?.Escape || 9, function () {
        if (document.activeElement instanceof HTMLElement) {
          document.activeElement.blur();
        }
      });
    }
    this.afterMonaco();
  }

  // TokenType match must be set smarter.
  public getTokens(tokenType: string = 'keyword', language = this.language) {
    // Dynamically loaded
    let monaco = (window as any).monaco;
    let tags = new Set<string>();
    if (monaco && this.descriptionControl) {
      let tokenArr = monaco.editor.tokenize(this.descriptionControl.value, language);
      let m = this.monacoEditor.getModel();

      // Custom tag matching per line I guess where we remove the matched tags from the loop but it might
      // actually be faster to just do a contains against the entire string.
      // The word offset boundry is all messed up...
      let currentLang = _.find(monaco.languages.getLanguages(), {
        id: language,
      });
      // console.log("currentLanguage", currentLang, monaco.languages.getLanguages());

      let match = `${tokenType}.${language}`;
      _.each(tokenArr, (tokens, lineIdx: number) => {
        if (!tokens) return;
        let line;
        try {
          line = m.getLineContent(lineIdx + 1);
        } catch (e) {
          // console.error("A delete can trigger an event and the model updates under you", lineIdx);
        }
        if (line === undefined) return;

        _.each(tokens, token => {
          //console.log("token.type", token.type)
          if (token.type == match) {
            // This should work, but doesn't because of crazy word boundry monaco stuff?
            // let position = m.getPositionAt(token.offset + 1);
            // position.lineNumber = lineIdx + 1;
            // let word = m.getWordAtPosition(position);

            // The highlights work but the word positions are all jacked up
            let word = this.readToken(line, token.offset);
            if (this.tagLookup[word]) {
              tags.add(word);
            } else {
              this.processProblemTags(line, tags);
            }
          }
        });
      });
    }
    return Array.from(tags);
  }

  processProblemTags(line: string, tags: Set<string>) {
    // console.log("Problem line", line);
    _.each(this.problemTags, t => {
      if (line.includes(t.id)) {
        tags.add(t.id);
      }
    });
  }

  // Additional keys for token matching?
  readToken(str: string, offset: number) {
    let code, j;
    let len = str.length;
    if (offset > len) {
      return '';
    }
    for (j = offset; j < len; ++j) {
      code = str.charCodeAt(j);
      if (
        !(code > 47 && code < 58) && // numeric (0-9)
        !(code > 64 && code < 91) && // upper alpha (A-Z)
        !(code === 95) && // upper alpha (A-Z)
        !(code > 96 && code < 123)
      ) {
        // lower alpha (a-z)
        break;
      }
    }
    return str.slice(offset, j);
  }

  public afterMonaco() {
    console.log('After monaco initialization.');
    if (!this.editForm) {
      return;
    }

    // Subscribes specifically to the description changes.
    let control = this.editForm.get('description');
    if (control) {
      control.valueChanges.pipe(distinctUntilChanged(), debounceTime(500)).subscribe({
        next: (evt: any) => {
          if (this.editForm) {
            // console.log("VS Code editor form updated", this.editForm.value);
          }
        },
      });
      // Set this after the initialization.
      this.editorValue = control.value;
    }
    _.delay(() => {
      this.fitContent();
    }, 50);

    _.delay(() => {
      this.createPlaceholder(this.placeholder, this.monacoEditor);
    }, 200);

    _.delay(() => {
      this.changeEmitter.emit({
        tags: this.getTokens(),
        value: this.editorValue,
      });
    }, 20);
  }

  createPlaceholder(placeholder: string, editor: MonacoEditor.ICodeEditor) {
    // Need to make it so the placeholder cannot be clicked
    //console.log("Placeholder", placeholder, editor, "TS Wrapper", this.editor);
    new PlaceholderContentWidget(this.placeholder, editor);
  }

  // TODO:  This also needs to handle a window resize event to actually check the content
  // and do a redraw. Might also be better to hide till the first redraw event.
  // https://github.com/microsoft/monaco-editor/issues/568
  fitContent() {
    let el = this.container.nativeElement;
    let width = el.offsetWidth;

    let updateHeight = () => {
      let editor = this.monacoEditor;
      let lineCount = this.fixedLineCount;
      if (lineCount < 1) {
        lineCount = Math.max(editor.getModel()?.getLineCount(), 8);
      }

      // You would think this would work but unfortunately the height of content is altered
      // by the spacing of the render so it expands forever.
      //const contentHeight = Math.min(2000, this.monacoEditor.getContentHeight());
      let contentHeight = 19 * (lineCount + this.padLine);
      el.style.height = `${contentHeight}px `;
      el.style.width = `${width}px `;
      editor.layout({ width, height: contentHeight });
    };

    // Already delayed after the initialization
    updateHeight();

    let changed = _.debounce(updateHeight, 150);
    this.monacoEditor.onDidChangeModelDecorations(changed);
  }
}

/*
 * Represents an placeholder renderer for monaco editor
 * Roughly based on https://github.com/microsoft/vscode/blob/main/src/vs/workbench/contrib/codeEditor/browser/untitledTextEditorHint/untitledTextEditorHint.ts
 */
class PlaceholderContentWidget implements MonacoEditor.IContentWidget {
  private static readonly ID = 'editor.widget.placeholderHint';

  private domNode: HTMLElement | undefined;

  constructor(
    private readonly placeholder: string,
    private readonly editor: MonacoEditor.ICodeEditor
  ) {
    // register a listener for editor code changes
    editor.onDidChangeModelContent(() => this.onDidChangeModelContent());
    // ensure that on initial load the placeholder is shown
    this.onDidChangeModelContent();
  }

  private onDidChangeModelContent(): void {
    if (this.editor.getValue() === '') {
      this.editor.addContentWidget(this);
    } else {
      this.editor.removeContentWidget(this);
    }
  }

  getId(): string {
    return PlaceholderContentWidget.ID;
  }

  getDomNode(): HTMLElement {
    if (!this.domNode) {
      this.domNode = document.createElement('div');
      this.domNode.style.width = 'max-content';
      this.domNode.style.pointerEvents = 'none';
      this.domNode.textContent = this.placeholder; // Could update with image
      this.domNode.style.fontStyle = 'italic';
      this.editor.applyFontInfo(this.domNode);
    }

    return this.domNode;
  }

  getPosition(): MonacoEditor.IContentWidgetPosition | null {
    // The whole typing import vs loading async via M$ is super messy.
    const editor = MonacoEditor || (document as any).monaco?.editor;

    return {
      position: { lineNumber: 1, column: 1 },
      preference: [editor?.ContentWidgetPositionPreference?.EXACT],
    };
  }

  dispose(): void {
    this.editor.removeContentWidget(this);
  }
}
