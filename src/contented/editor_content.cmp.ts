/**
 * Provide the ability to edit the descriptions of content and containers.  Also provide the ability
 * to quickly manage tags.  TODO: This component should actually be broken into a pure wrapper around
 * the ngx-monaco intialization and handle just readonly and change emitting.
 */
import { Component, OnInit, Input, EventEmitter, ViewChild } from '@angular/core';
import { ActivatedRoute, ParamMap } from '@angular/router';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { finalize } from 'rxjs/operators';
import { ContentedService } from './contented_service';
import { Tag, Content, VideoCodecInfo } from './content';
import { VSCodeEditorCmp } from './vscode_editor.cmp';
import { TaskOperation, TaskRequest } from './task_request';
import { GlobalBroadcast } from './global_message';
import { Screen } from './screen';
import * as _ from 'lodash';

@Component({
    selector: 'editor-content-cmp',
    templateUrl: './editor_content.ng.html',
    standalone: false
})
export class EditorContentCmp implements OnInit {
  @ViewChild('description') editor: VSCodeEditorCmp | undefined;

  @Input() content?: Content;

  @Input() editForm: FormGroup;
  @Input() descriptionControl: FormControl = new FormControl('', Validators.required);

  @Input() screensForm: FormGroup;
  @Input() offsetControl: FormControl<number | null> = new FormControl(null, Validators.required);
  @Input() countControl: FormControl<number | null> = new FormControl(12, Validators.required);
  @Input() checkStates = true;

  public taskCreated: EventEmitter<any> = new EventEmitter<any>();
  // These are values for the Monaco Editors, change events are passed down into
  // the form event via the AfterInit and set the v7_definition & suricata_definition.
  public loading: boolean = false;
  public taskLoading: boolean = false;

  // Mostly we use format.duration
  public vidInfo: VideoCodecInfo | undefined;

  constructor(
    public fb: FormBuilder,
    public route: ActivatedRoute,
    public _service: ContentedService
  ) {
    this.editForm = this.fb.group({
      description: this.descriptionControl,
    });

    this.screensForm = this.fb.group({
      offset: this.offsetControl,
      count: this.countControl,
    });
  }

  // Subscribe to options changes, if the definition changes make the call
  public ngOnInit() {
    if (!this.content) {
      this.route.paramMap.pipe().subscribe({
        next: (map: ParamMap) => {
          console.log('Reloading content');
          this.content = undefined; // Changing the content will force a reload
          this.loadContent(map.get('id') || '');
        },
        error: err => {
          GlobalBroadcast.error('Loading content for editing', err);
        },
      });
    }
  }

  loadContent(id: string) {
    this._service.getContent(id).subscribe({
      next: (content: Content) => {
        this.content = content;
        this.descriptionControl.setValue(content.description);
        if (content.isVideo()) {
          this.vidInfo = content.getVideoInfo();
        }
      },
      error: err => {
        GlobalBroadcast.error(`Could not load ${id}`, err);
      },
    });
  }

  save() {
    if (!this.content || !this.editForm) {
      return;
    }
    console.log('Save()', this.editForm.value);
    this.content.description = _.get(this.editForm.value, 'description');
    this.loading = true;

    let tags = this.editor?.getTokens() || [];
    this.content.tags = _.map(tags, tag => new Tag(tag));
    this._service
      .saveContent(this.content)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: ret => {
          console.log('Saved content', ret);
        },
        error: err => {
          GlobalBroadcast.error('Could not save changes', err);
        },
      });
  }

  clearScreens(content: Content) {
    if (!content) {
      return;
    }
    this._service.clearScreens(content.id).subscribe({
      next: (content: Content) => {
        this.content = content;
      },
      error: err => {
        GlobalBroadcast.error('Could not clear screens', err);
      },
    });
  }

  // Generate incremental screens and then check the request
  incrementalScreens(content: Content) {
    if (!content || !this.screensForm) {
      return;
    }
    let req = this.screensForm.value;
    this.taskLoading = true;
    this._service
      .requestScreens(content, req.count, req.offset)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Success requesting new task for content', task, content);
          this.watchTask(task);
        },
        error: err => {
          GlobalBroadcast.error('Failed to get new screens task', err);
        },
      });
  }

  canReEncode(content: Content) {
    if (!this.taskLoading && content && content.isVideo()) {
      let info = content.getVideoInfo();
      return info ? info.CanEncode : false;
    }
    return false;
  }

  encodeVideoContent(content: Content) {
    this.taskLoading = true;
    this._service
      .encodeVideoContent(content)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Created video encoding task', content, task);
          this.watchTask(task);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start encoding tasks', err);
        },
      });
  }

  tagContent(content: Content) {
    this.taskLoading = true;
    this._service
      .createTagContentTask(content)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Created a tagging task', content, task);
          this.watchTask(task);
        },
        error: err => {
          GlobalBroadcast.error('Failed to start tagging tasks', err);
        },
      });
  }

  createPreviewFromScreens(content: Content) {
    console.log('Create a preview');
    this.taskLoading = true;
    this._service
      .createPreviewFromScreens(content)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Created preview screen tasks', content, task);
          this.watchTask(task);
        },
        error: err => {
          GlobalBroadcast.error('Failed to kick off preview task', err);
        },
      });
  }

  findDupesContent(content: Content) {
    console.log('Create dupes');
    this.taskLoading = true;
    this._service
      .findDuplicateForContentTask(content)
      .pipe(finalize(() => (this.taskLoading = false)))
      .subscribe({
        next: (task: TaskRequest) => {
          console.log('Find dupes', content, task);
          this.watchTask(task);
        },
        error: err => {
          GlobalBroadcast.error('Failed to kick off dupe task', err);
        },
      });
  }

  taskUpdated(task: TaskRequest) {

    console.log('Task updated', task, task.operation);
    const contentId = this.content?.id;
    if (!this.content || !contentId) {
      return;
    }

    if (task.operation === TaskOperation.SCREENS) {
      this.content.screens = [];
      this.loadScreens(contentId);
    }

    if (task.operation === TaskOperation.WEBP) {
      this.content = undefined;
      this.loadContent(contentId);
    }
  }

  loadScreens(contentId: string) {
    this._service.getScreens(contentId).subscribe({
      next: (screens: { total: number; results: Array<Screen> }) => {
        if (!this.content) {
          return;
        }
        this.content.screens = screens.results;
      },
      error: err => {
        GlobalBroadcast.error('Failed to load screens', err);
      },
    });
  }

  canCreatePreview(content: Content) {
    if (!this.taskLoading && content && content.screens && content.screens.length > 0) {
      return true;
    }
    return false;
  }

  // Start watching the task queue
  watchTask(task: TaskRequest) {
    this.taskCreated.emit(task);
  }
}
