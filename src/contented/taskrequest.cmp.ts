import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {finalize, debounceTime, distinctUntilChanged} from 'rxjs/operators';
import {ContentedService} from './contented_service';
import {TaskRequest, TASK_STATES} from './task_request';
import { MatTableDataSource} from '@angular/material/table';
// import {ActivatedRoute, Router, ParamMap} from '@angular/router';
import {FormBuilder, NgForm, FormControl, FormGroup} from '@angular/forms';

import * as _ from 'lodash-es';

@Component({
  selector: 'task-request-cmp',
  templateUrl: './taskrequest.ng.html',
})
export class TaskRequestCmp implements OnInit {

    @Input() contentID: string = "";
    @Input() pageSize = 100;
    @Input() reloadEvt: EventEmitter<any>; // Do you want to reload the task queue
    @Output() taskUpdated: EventEmitter<TaskRequest> = new EventEmitter<TaskRequest>;
    @Input() checkStates = false;

    public loading = false;
    public tasks: Array<TaskRequest>;
    public total = 0;

    displayedColumns: string[] = ['operation', 'status', 'actions', 'created_at', 'updated_at', 'message', 'created_id', 'error'];
    dataSource = new MatTableDataSource<TaskRequest>([]);
    states = TASK_STATES;

    searchForm: FormGroup;
    status: FormControl<string> = new FormControl("");
    search: FormControl<string> = new FormControl("");

    //constructor(public _service: ContentedService, public route: ActivatedRoute) {
    constructor(public _service: ContentedService, fb: FormBuilder) {
      this.searchForm = fb.group({
        search: this.search,
        status: this.status,
      });
    }

    ngOnInit() {
      this.loadTasks(this.contentID);

      if (this.reloadEvt) {
        this.reloadEvt.subscribe(() => {
          _.delay(() => {
            this.loadTasks(this.contentID);
          }, 2000);
        }, console.error)
      }

      this.searchForm.valueChanges.pipe(
        debounceTime(500),
        distinctUntilChanged()
        // Prevent bubble on keypress
      ).subscribe({
          next: (formData) => {
              return this.loadTasks(this.contentID, [], formData.status, formData.search)
          },
          error: error => {
            console.error("Failed to search Tasks error", error);
          }
      });
      if (this.checkStates) {
        this.pollStart();
      }
    }

    loadTasks(contentID: string, watching: Array<TaskRequest> = [], status = "", search = "") {
      this.loading = true;
      return this._service.getTasks(this.contentID, 1, this.pageSize, status, search).pipe(
        finalize(() => this.loading = false)
      ).subscribe(
        (taskResponse) => {
          this.tasks = taskResponse.results;
          this.total = taskResponse.total;
          this.dataSource = new MatTableDataSource<TaskRequest>(this.tasks || [])

          this.checkComplete(this.tasks, watching);
        },
        console.error
      );
    }

    cancelTask(task: TaskRequest) {
      console.log("Attempt to cancel task", task);
      task.uxLoading = true;
      this._service.cancelTask(task).pipe(
        finalize(() => task.uxLoading = false)
      ).subscribe({
        next: (taskResponse) => {
          return new TaskRequest(taskResponse);  
        },
        error: (err) => console.error
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
        this.pollStart()
      }, 1000 * 10);
    }

    pollTasks() {
      if (this.loading) {
        return
      }
      let notComplete: Array<TaskRequest> = _.filter(this.tasks, task => {
        return task.isComplete();
      }) || [];
      let vals = this.searchForm.value;
      this.loadTasks(this.contentID, notComplete, vals.status, vals.search);
    }

    pageEvt(evt: any) {
      console.log("Page event is annoying", evt);
    }
}
