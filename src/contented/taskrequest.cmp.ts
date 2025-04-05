import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { ContentedService, TaskSearch, TaskStatus } from './contented_service';
import { TaskRequest, TASK_STATES } from './task_request';
import { MatTableDataSource } from '@angular/material/table';
// import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { GlobalBroadcast } from './global_message';
import { Subscription } from 'rxjs';
import { initializeDefaults } from './utils';
import { createStringControl } from './form_utils';

import * as _ from 'lodash';

@Component({
    selector: 'task-request-cmp',
    templateUrl: './taskrequest.ng.html',
    standalone: false
})
export class TaskRequestCmp implements OnInit {
  @Input() contentID: string = '';
  @Input() pageSize = 100;
  @Input() reloadEvt!: EventEmitter<any>; // Do you want to reload the task queue
  @Output() taskUpdated: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>();
  @Input() checkStates = false;

  public loading = false;
  public tasks: Array<TaskRequest> = [];
  public total = 0;
  public statusFilterControl: FormControl<string>;
  public searchFilterControl: FormControl<string>;
  public reloadSub: Subscription | undefined;
  public throttleSub: Subscription | undefined;
  public tasksForm: FormGroup;

  displayedColumns: string[] = [
    'operation',
    'status',
    'actions',
    'created_at',
    'started_at',
    'updated_at',
    'message',
    'created_id',
    'error',
  ];
  dataSource = new MatTableDataSource<TaskRequest>([]);
  states = TASK_STATES;

  constructor(
    public _service: ContentedService,
    fb: FormBuilder
  ) {
    // Initialize default values
    initializeDefaults(this, {
      statusFilterControl: createStringControl(''),
      searchFilterControl: createStringControl('')
    });

    this.tasksForm = new FormGroup({
      status: this.statusFilterControl,
      search: this.searchFilterControl
    });
  }

  ngOnInit() {
    this.reload(true);
    if (this.reloadEvt) {
      this.reloadSub = this.reloadEvt.subscribe({
        next: () => {
          this.reload(true);
        }
      });
    }

    this.throttleSub = this.tasksForm.valueChanges
      .pipe(debounceTime(350), distinctUntilChanged())
      .subscribe({
        next: res => {
          this.reload(false);
        }
      });

    // If it should be checking for completed tasks, start polling, vs just load it up once for a state check
    if (this.checkStates) {
      this.pollStart();
    } else {
      this.loadTasks(this.contentID);
    }
  }

  reload(initial: boolean) {
    const status = this.statusFilterControl.value as TaskStatus;
    this.loadTasks(this.contentID, [], status, this.searchFilterControl.value);
  }

  loadTasks(contentID: string, notComplete: Array<TaskRequest> = [], status: TaskStatus = '', search = '') {
    this.loading = true;

    const query: TaskSearch = {
      contentID,
      status,
      search,
      offset: 0,
      limit: this.pageSize,
    };

    console.log('Loading tasks', query, notComplete);
    return this._service
      .getTasks(query)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe({
        next: taskResponse => {
          // On an initial load we need to get the not complete tasks and don't want events
          // for tasks completed long ago.
          this.tasks = taskResponse.results;
          this.total = taskResponse.total;
          this.dataSource = new MatTableDataSource<TaskRequest>(this.tasks || []);

          this.checkComplete(this.tasks, notComplete);
        },
        error: err => {
          GlobalBroadcast.error('Failed to load tasks', err);
        },
      });
  }

  cancelTask(task: TaskRequest) {
    console.log('Attempt to cancel task', task);
    task.uxLoading = true;
    this._service
      .cancelTask(task)
      .pipe(finalize(() => (task.uxLoading = false)))
      .subscribe({
        next: taskResponse => {
          return new TaskRequest(taskResponse);
        },
        error: err => console.error,
      });
  }

  checkComplete(tasks: Array<TaskRequest>, watching: Array<TaskRequest> = []) {
    let check = _.keyBy(watching, 'id');
    (tasks || []).forEach(task => {
      if (check[task.id] && task.isComplete()) {
        this.taskUpdated.emit(task);
      }
    });
  }

  pollStart() {
    this.pollTasks();
    _.delay(() => {
      this.pollStart();
    }, 1000 * 10);
  }

  pollTasks() {
    if (this.loading) {
      return;
    }
    let watching: Array<TaskRequest> = _.filter(this.tasks, task => !task.isComplete()) || [];
    let vals = this.tasksForm.value;
    this.loadTasks(this.contentID, watching, vals.status, vals.search);
  }

  pageEvt(evt: any) {
    console.log('Page event is annoying to handle', evt);
  }
}
