import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';
import {HttpClientTestingModule, HttpTestingController} from '@angular/common/http/testing';
import {By} from '@angular/platform-browser';
import {DebugElement} from '@angular/core';

import {ContentedModule} from '../contented/contented_module';
import { TaskRequestCmp } from './taskrequest.cmp';
import {MockData} from '../test/mock/mock_data';
import { RouterTestingModule } from '@angular/router/testing';

declare var $;

describe('TaskRequestCmp', () => {
  let component: TaskRequestCmp;
  let fixture: ComponentFixture<TaskRequestCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        NoopAnimationsModule,
        ContentedModule,
        HttpClientTestingModule,
        RouterTestingModule.withRoutes(
          [{path: 'admin_ui/tasks', component: TaskRequestCmp}]
        ),
      ],
      declarations: [TaskRequestCmp]
    });
    fixture = TestBed.createComponent(TaskRequestCmp);
    component = fixture.componentInstance;
    fixture.detectChanges();

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.task-request-cmp'));
    el = de.nativeElement;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  fit('Should be trying to load tasks', fakeAsync(() => {
    fixture.detectChanges();

    let req = httpMock.expectOne(req => req.url.includes('/task_requests'));
    req.flush(MockData.taskRequests());
    tick(1000);

    expect(component.tasks?.length).withContext("The tasks should be set").toEqual(6);

    fixture.detectChanges();
    expect($(".task-operation").length).withContext("Render the tasks.").toEqual(6)
    tick(1000);
  }));
});
