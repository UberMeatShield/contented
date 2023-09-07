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
import {Container} from './container';

import * as _ from 'lodash-es';

import {RESUME} from './resume';

@Component({
  selector: 'splash-cmp',
  templateUrl: './splash.ng.html',
})
export class SplashCmp implements OnInit {

  @ViewChild('EDITOR') editor?: EditorComponent;

  @Input() editForm?: FormGroup;
  @Input() editorValue: string = RESUME; // TODO: Save this as media
  @Input() descriptionControl?: FormControl<string>;
  @Input() readOnly: boolean = true;
  @Input() editorOptions = {
    //theme: 'vs-dark',
    //language: 'html',
    language: 'tagging',
  };
  @Input() mc?: Content;
  @Input() c?: Container;
  @Input() splashTitle: string = "";
  @Input() splashContent: string = "";
  @Input() rendererType: string = "";

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
    if (!this.mc && !this.c) {
      this.loadSplash();
    }
  }

  // Load the splash page instead of a particular content id
  loadSplash() {
      //this.loading = true;
      console.log("Load splash media content");
      this._service.splash().subscribe(
        res => {
          this.c = res.container;
          this.mc = res.content;
          this.splashTitle = res.splashTitle || "";
          this.splashContent = res.splashContent || "";
          this.rendererType = res.rendererType;
        },
         console.error
      );
  }

  getVideos() {
    if (!this.c) {
      return []
    }
    return _.filter(this.c.contents, mc => {
        return mc.content_type.includes("video");
    });
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
    this.fitContent();
  }

  fitContent() {
    console.log("Content fit");
    const container = document.getElementById('SPLASH_FULL');
    let width = container.offsetWidth;
    //container.style.border = '1px solid #f00';

    const updateHeight = () => {
    	const contentHeight = Math.min(9000, this.monacoEditor.getContentHeight());
    	container.style.width = `${width}px`;
    	container.style.height = `${contentHeight}px`;
    	this.monacoEditor.layout({width, height: contentHeight });
    };

    // Give other page elements a little bit of rendering time
    _.delay(() => {
      updateHeight();
    }, 100);

    let changed = _.debounce(updateHeight, 250);
    this.monacoEditor.onDidContentSizeChange(changed);
  }
}
