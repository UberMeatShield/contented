/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup, Validators} from '@angular/forms';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {MatRipple} from '@angular/material/core';
import {EditorComponent} from 'ngx-monaco-editor-v2';
import {ContentedService} from './contented_service';
import {Content} from './content';

import * as _ from 'lodash-es';

@Component({
  selector: 'media-editor-cmp',
  templateUrl: './media_editor.ng.html',
})
export class MediaEditorCmp implements OnInit {

  @ViewChild('EDITOR') editor?: EditorComponent;

  @Input() editForm?: FormGroup;
  @Input() editorValue: string = "";
  @Input() descriptionControl?: FormControl<string>;
  @Input() readOnly: boolean = false;
  @Input() editorOptions = {
    theme: 'vs-dark',
    //language: 'html',
    language: 'tagging',
  };
  @Input() mc?: Content;

  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  @Output() changeEmitter = new EventEmitter<string>();
  public loading: boolean = false;

  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: any;


  constructor(public fb: FormBuilder, public route: ActivatedRoute, public _service: ContentedService) {
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.editForm) {
      this.editForm = this.fb.group({
        "description": this.descriptionControl = (this.descriptionControl || new FormControl(this.editorValue || "")),
      });
    }
    if (!this.mc) {
        this.route.paramMap.pipe().subscribe(
            (map: ParamMap) => {
                this.loadContent(map.get('id'));
            },
            console.error
        );
    }
  }

  loadContent(id: string) {
      this._service.getContent(id).subscribe(
          (mc: Content) => {
              console.log(mc);
              this.mc = mc;
              this.descriptionControl.setValue(mc.description);
          },
          console.error
      )
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

  afterMonacoInit(monaco: any) {
    console.log("Monaco Editor has been initialized");
    this.monacoEditor = monaco;
      // This is a little awkward but we need to be able to change the form control
    if (this.editor) {
      this.changeEmitter.pipe(
        distinctUntilChanged(),
        debounceTime(250)
      ).subscribe(
        (val: string) => {
            console.log("Changed", val);
        },
        console.error
      );

      this.editor.registerOnChange((val: string) => {
        this.changeEmitter.emit(val);
      });
    }
    this.afterMonaco();
    this.setReadOnly(this.readOnly);
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
        debounceTime(1000)
      ).subscribe(
        (evt: any) => {
          if (this.editForm) {
            console.log(this.editForm);
          }
        }
      );
      this.editorValue = control.value;
    }
  }
}
