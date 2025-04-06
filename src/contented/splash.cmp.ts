/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.
 */
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild } from '@angular/core';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { FormBuilder, NgForm, FormControl, FormGroup, Validators } from '@angular/forms';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { MatRipple } from '@angular/material/core';
import { EditorComponent } from 'ngx-monaco-editor-v2';
import { ContentedService } from './contented_service';
import { Content } from './content';
import { Container } from './container';

import * as _ from 'lodash';

import { RESUME } from './resume';

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
  @Input() splashTitle: string = '';
  @Input() splashContent: string = '';
  @Input() rendererType: string = '';

  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  @Output() changeEmitter = new EventEmitter<string>();
  public loading: boolean = false;

  // Reference to the raw Microsoft component, allows for
  public monacoEditor?: any;

  constructor(
    public fb: FormBuilder,
    public route: ActivatedRoute,
    public _service: ContentedService
  ) {}

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.editForm) {
      this.editForm = this.fb.group({
        description: (this.descriptionControl = this.descriptionControl || new FormControl(this.editorValue || '')),
      });
    }
    if (!this.mc && !this.c) {
      this.loadSplash();
    }
  }

  // Load the splash page instead of a particular content id
  loadSplash() {
    console.log('Load splash media content');
    this.loading = true;
    this._service
      .splash()
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: res => {
          this.c = res.container;
          this.mc = res.content;
          this.splashTitle = res.splashTitle || '';
          this.splashContent = res.splashContent || '';
          this.rendererType = res.rendererType;
        },
        error: error => console.error(error),
      });
  }

  getVideos() {
    return _.filter(this.c?.contents || [], mc => {
      return mc.content_type.includes('video');
    });
  }
}
