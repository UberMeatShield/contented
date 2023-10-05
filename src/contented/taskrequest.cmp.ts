import { Component, OnInit, AfterViewInit, Input, Output, EventEmitter, ViewChild} from '@angular/core';
import {ContentedService} from './contented_service';

@Component({
  selector: 'task-request-cmp',
  templateUrl: './taskrequest.ng.html',
})
export class TaskRequestCmp implements OnInit {

    @Input() contentID: string = "";

    constructor(public _service: ContentedService) {
    }

    ngOnInit() {
      console.log("init task request");     
      this._service.getTasks(this.contentID, 1, 0).subscribe(
        console.log, console.error
      )
    }
}
