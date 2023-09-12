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
    console.log("Monaco Editor has been initialized");
    this.monacoEditor = monaco;
      // This is a little awkward but we need to be able to change the form control
    if (this.editor) {
      this.changeEmitter.pipe(
        distinctUntilChanged(),
        debounceTime(50)
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
    //this.setReadOnly(this.readOnly);
  }

  public afterMonaco() {
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
          console.log("Control updated", evt);
          if (this.editForm) {
            console.log("VS Code editor form updated", this.editForm.value);
          }
        }
      );
      this.editorValue = control.value;
    }

    _.delay(() => {
      this.fitContent();
    }, 500);
  }

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