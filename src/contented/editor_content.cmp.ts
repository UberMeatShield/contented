/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.  TODO: This component should actually be broken into a pure wrapper around
 * the ngx-monaco intialization and handle just readonly and change emitting.
 */
import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, FormControl, FormGroup, Validators} from '@angular/forms';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {Tag, Content, VideoCodecInfo} from './content';
import {VSCodeEditorCmp} from './vscode_editor.cmp';

import * as _ from 'lodash-es';

@Component({
  selector: 'editor-content-cmp',
  templateUrl: './editor_content.ng.html',
})
export class EditorContentCmp implements OnInit {

  @ViewChild('description') editor: VSCodeEditorCmp;

  @Input() content?: Content;

  @Input() editForm?: FormGroup;
  @Input() descriptionControl: FormControl = new FormControl("", Validators.required);

  @Input() screensForm?: FormGroup;
  @Input() offsetControl: FormControl<number>; 
  @Input() countControl: FormControl<number>;

  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  public loading: boolean = false;
  public taskLoading: boolean = false;

  // Mostly we use format.duration
  public vidInfo: VideoCodecInfo;

  constructor(public fb: FormBuilder, public route: ActivatedRoute, public _service: ContentedService) {
    this.editForm = this.fb.group({
      description: this.descriptionControl
    });

    this.offsetControl = new FormControl(undefined, Validators.required)
    this.countControl = new FormControl(12, Validators.required)
    this.screensForm = this.fb.group({
      offset: this.offsetControl,
      count: this.countControl,
    })
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.content) {
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
          (content: Content) => {
              this.content = content;
              this.descriptionControl.setValue(content.description);
              if (content.isVideo()) {
                this.vidInfo = content.getVideoInfo();
              }
          },
          console.error
      )
  }

  save() {
    console.log("Save()", this.editForm.value);
    this.content.description = _.get(this.editForm.value, 'description');
    this.loading = true;

    let tags = this.editor.getTokens();
    console.log(tags);
    this.content.tags = _.map(tags, tag => new Tag(tag));
    this._service.saveContent(this.content).pipe(finalize(() => this.loading = false)).subscribe(
      console.log,
      console.error
    );
  }

  // Generate incremental screens and then check the request
  incrementalScreens(content) {
    let req = this.screensForm.value;
    this.taskLoading = true;
    this._service.requestScreens(content, req.count, req.offset).pipe(finalize(() => this.taskLoading = false)).subscribe(
      console.log,
      console.error
    )
  }

  canReEncode(content: Content) {
    if (!this.taskLoading && content && content.isVideo()) {
      let info = content.getVideoInfo()
      return info ? info.CanEncode : false;
    }
    return false
  }

  encodeVideoContent(content: Content) {
    this.taskLoading = true; 
    this._service.encodeVideoContent(content).pipe(finalize(() => this.taskLoading = false)).subscribe(
      console.log,
      console.error
    )
  }
}
