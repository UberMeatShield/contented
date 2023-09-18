/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild, ElementRef} from '@angular/core';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup, Validators} from '@angular/forms';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {MatRipple} from '@angular/material/core';
import {EditorComponent} from 'ngx-monaco-editor-v2';
import {ContentedService} from './contented_service';
import {Content} from './content';
import {Container} from './container';

import * as _ from 'lodash-es';
import * as $ from 'jquery';

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

  @Input() editorOptions = {
    theme: 'vs-dark',
    language: this.language,
  };
  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  @Output() changeEmitter = new EventEmitter<string>();


  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: any;
  public initialized = false;

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

    this.monacoDigestHackery()
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
      ).subscribe(
        (val: string) => {
            this.editForm.get("description").setValue(val);
        },
        console.error
      );
      this.editor.registerOnChange((val: string) => {
        this.changeEmitter.emit(val);
      });
    }
    this.afterMonaco();
  }

  public getTokens(tokenType: string = "keyword", language = this.language) {
    // Dynamically loaded
    let monaco = (window as any).monaco;
    let tags = [];
    if (monaco && this.descriptionControl) {
      let tokenArr = monaco.editor.tokenize(this.descriptionControl.value, language);

      let m = this.monacoEditor.getModel();

      let match = `${tokenType}.${language}`;
      _.each(tokenArr, (line, lineIdx) => {
        _.each(line, token => {
          if (token.type == match) {
            // console.log(lineIdx + 1, token, m.getPositionAt(token.offset));
            let position = m.getPositionAt(token.offset);
            position.lineNumber = lineIdx + 1;

            let word = m.getWordAtPosition(position);
            tags.push(word.word);
          }
        })
      });
    }
    return _.uniq(_.compact(tags));
  }

  public afterMonaco() {
    console.log("After monaco");
    if (!this.editForm) {
      return;
    }

    // Subscribes specifically to the description changes.
    let control = this.editForm.get("description");
    if (control) {
      control.valueChanges.pipe(
        distinctUntilChanged(),
        debounceTime(500)
      ).subscribe(
        (evt: any) => {
          if (this.editForm) {
            // console.log("Control updated", evt);
            // console.log("VS Code editor form updated", this.editForm.value);
          }
        }
      );
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