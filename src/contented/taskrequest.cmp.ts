import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { finalize, debounceTime, distinctUntilChanged } from 'rxjs/operators';
import { ContentedService, TaskSearch, TaskStatus } from './contented_service';
import { TaskRequest, TASK_STATES } from './task_request';
import { MatTableDataSource } from '@angular/material/table';
// import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import { FormBuilder, FormControl, FormGroup } from '@angular/forms';
import { GlobalBroadcast } from './global_message';
import { PageEvent } from '@angular/material/paginator';

import * as _ from 'lodash';

interface TaskRequestForm {
  search: FormControl<string | null>;
  status: FormControl<string | null>;
}

@Component({
    selector: 'task-request-cmp',
    templateUrl: './taskrequest.ng.html',
    standalone: false
})
export class TaskRequestCmp implements OnInit {
  @Input() contentID: number = 0;
  @Input() pageSize = 100;
  @Input() reloadEvt: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>();
  @Output() taskUpdated: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>();
  @Input() checkStates = false;

  public loading = false;
  public tasks: TaskRequest[] = [];
  public total = 0;

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

  searchForm: FormGroup<TaskRequestForm>;
  status: FormControl<string | null> = new FormControl<string | null>('');
  search: FormControl<string | null> = new FormControl<string | null>('');

  //constructor(public _service: ContentedService, public route: ActivatedRoute) {
  constructor(
    public _service: ContentedService,
    fb: FormBuilder
  ) {
    this.searchForm = fb.group<TaskRequestForm>({
      search: this.search,
      status: this.status,
    });
  }

  ngOnInit() {
    if (this.reloadEvt) {
      this.reloadEvt.subscribe({
        next: (tr: TaskRequest) => {
          console.log('Reloading tasks', tr);
          _.delay(() => {
            try {
              const watched = [tr].concat(_.filter(this.tasks, task => !task.isComplete()));
              this.tasks = [];
              this.loadTasks(this.contentID, watched);
            } catch (err) {
              console.error('Failed to reload the tasks', err);
            }
          }, 1000);
        },
        error: (err: unknown) => {
          GlobalBroadcast.error('Failed to reload the tasks', err);
        },
      });
    }

    this.searchForm.valueChanges
      .pipe(
        debounceTime(500),
        distinctUntilChanged()
        // Prevent bubble on keypress
      )
      .subscribe({
        next: (formData: Partial<{ search: string | null; status: string | null }>) => {
          return this.loadTasks(this.contentID, [], (formData.status as TaskStatus) || '', formData.search || '');
        },
        error: (error: unknown) => {
          console.error('Failed to search Tasks error', error);
        },
      });

    // If it should be checking for completed tasks, start polling, vs just load it up once for a state check
    if (this.checkStates) {
      this.pollStart();
    } else {
      this.loadTasks(this.contentID);
    }
  }

  loadTasks(contentID: number, notComplete: TaskRequest[] = [], status: TaskStatus = '', search = ''): void {
    this.loading = true;

    const query: TaskSearch = {
      contentID: contentID.toString(),
      status,
      search,
      offset: 0,
      limit: this.pageSize,
    };

    console.log('Loading tasks', query, notComplete);
    this._service
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
        error: (err: unknown) => {
          GlobalBroadcast.error('Failed to load tasks', err);
        },
      });
  }

  cancelTask(task: TaskRequest): void {
    console.log('Attempt to cancel task', task);
    task.uxLoading = true;
    this._service
      .cancelTask(task)
      .pipe(finalize(() => (task.uxLoading = false)))
      .subscribe({
        next: taskResponse => {
          return new TaskRequest(taskResponse);
        },
        error: (err: unknown) => console.error,
      });
  }

  checkComplete(tasks: TaskRequest[], watching: TaskRequest[] = []): void {
    const check = _.keyBy(watching, 'id');
    (tasks || []).forEach(task => {
      if (check[task.id] && task.isComplete()) {
        this.taskUpdated.emit(task);
      }
    });
  }

  pollStart(): void {
    this.pollTasks();
    _.delay(() => {
      this.pollStart();
    }, 1000 * 10);
  }

  pollTasks(): void {
    if (this.loading) {
      return;
    }
    const watching: TaskRequest[] = _.filter(this.tasks, task => !task.isComplete()) || [];
    const vals = this.searchForm.value;
    this.loadTasks(this.contentID, watching, (vals.status as TaskStatus) || '', vals.search || '');
  }

  pageEvt(evt: PageEvent): void {
    console.log('Page event is annoying to handle', evt);
  }
}
